# HANDOFF — 2026-08-10 (midday). 235's fix is PROVEN and 8 of 11 logos are repaired. A fleet incident happened in the middle. COLD-START HERE.

Supersedes `HANDOFF_2026-08-09d_…` (whose "what remains" list is now largely done,
and whose site list was **wrong in two places** — see below). Read
`NOTES_209…` (08-10 section) for evidence and every misstep.
Contribute INTO the bug files — shared accounts.

## State

| item | state |
|---|---|
| `bugs_open/209` | defect fixed and live since v1.0.1276. Open only for **Phase 3** (retire the dead `{purpose}_uri` writers) |
| `bugs_open/235` fix candidate 1 (mig 360) | **BEHAVIOURALLY PROVEN 08-10** — first correct logo the brand branch has ever produced |
| `bugs_open/235` fix candidate 2 (the artefacts) | **8 of 11 repaired and verified at the served artefact.** 3 blocked, each needing an OWNER CALL |
| re-render (so pages actually show it) | gamesdesign **DONE end-to-end**; 7 sites fanned out, ~250 page items draining at ~36/hour |
| stale `logo.jpg` deletion | **NOT started, deliberately** — must come last |
| `bugs_open/240` (NEW) | fleet incident, ended, **stays OPEN** — the sweep bought ~a week |
| `bugs_open/231` fleet census | still undone (61 `ActionInputSpec`s with Defaults) |
| `bugs_open/236` | still unowned |

## 1. What is DONE, and how it was proven

**The brand branch works.** One item at cookly.uk; the run recorded its own routing:
`check_imagery_brand_update = {condition_met:true, next_step_override:
store_imagery_brand_asset}`, `asset_stored.purpose = 'logo'`. Pre-360 that input
produced `hero`.

**8 sites repaired**, verified at the SERVED artefact, not the item status:
dartsonline, fundamentallyai, gamesdesign, lendzy, oufe, vetcomparison, vonc,
webdesign.co.uk — all now serve a PNG within a 400px box where several previously
404'd on `.png` entirely. `assets.purpose='logo'` on all 8.

## 2. Three corrections to the previous handoff — check these before trusting anything else it said

1. **The acceptance bar "400×400 PNG" is wrong and would FAIL a correct run.**
   Logo processing fits a 400px box preserving aspect: gamesdesign came out
   **400×400**, everything else **400×218**. Assert **PNG-not-JPEG**, **≤400px**,
   and **`purpose='logo'`** on the stamped row.
2. **No new `assets` row is created** — `store_asset` UPDATES in place. "Assert
   the NEW row" finds nothing.
3. **The site list was wrong twice.** `idea.uk` is **NOT** affected (correct PNG,
   pages reference it, only a stale unreferenced `logo.jpg`). `relojistas.com`
   **IS** affected and was missing from the list.

## 3. The three sites still wrong — each needs a DECISION, not a retry

- **`robot-hands.com` — permanent approval lock.** `locked_by=user-b6-approval`,
  2026-07-11. `store_asset` refused; the deploy then died on
  `asset_stored.image_uri not found`. **Do not retry.** The locked asset IS the
  defective one, but the lock protects approved *artwork*; regenerating discards
  it. The defect is encoding/size, not subject — so the repair is **re-deploying
  the EXISTING source object at `purpose='logo'`**, which is a different
  operation from the one the other eight ran.
- **`relojistas.com`** — also owner-locked (`bugs_open/131`) **and**
  `github_repo='vm-sites'`, a deploy path this lane has not verified reaches a
  VM-served site. Excluded deliberately.
- **`webdesign.uk`** — unlocked, but **blocked from ALL dispatch**: work item
  `8793da9a` was inserted directly as `status='claimed'` with NULL `claimed_at`,
  and the selector skips any site holding a claimed item. Clear that row and it
  repairs like the other eight (subject to the same `vm-sites` question). This is
  starving that site of every kind of build work, not just this one.

`[MEASURED]` **no other affected site carries a lock** — checked across all 11, so
nothing owner-approved was overwritten by this run.

## 4. What remains on the logo work, IN THIS ORDER

1. **Let the ~250 page re-renders drain** (~36/hour measured). They are queued and
   self-driving. Verify by **grepping the served HTML**, never by item status.
2. **Only then delete the stale `logo.jpg`** on the repaired sites. Nothing
   deletes it automatically and the pages reference it until re-rendered — delete
   early and you break the site.
3. The three sites in §3, after their owner decisions.

## 5. Traps paid for THIS session — do not re-derive

- **A `page_rerender` cannot change the logo/header/head.** Those live in
  `site_components`; a page re-render assembles from the EXISTING ones. It
  completes, advances `deployed_at`, and produces a byte-identical page. Only the
  site-level `needs_rerender` (`{"refresh_site_components": true}`) refreshes them
   — and even that only QUEUES the per-page items. **Three attempts, first two
  both "succeeded".** Full entry in `LANDMINES.md`.
- **Byte size is not a change detector here** — `logo.jpg` and `logo.png` are the
  same length, so a correct fix leaves the page exactly the same size. Grep the
  reference.
- **Filing a work item at the discovery default `detected` goes NOWHERE.**
  `build-pipeline-trigger`'s pre_query requires exactly `status='triaged'` AND
  `pipeline='build'`. Nothing drives `triage_detected_items`.
- **`kafka-topics.sh --list` truncates when piped** — down a `kubectl exec` stream
  AND inside the pod. Redirect to a file, process the file, read the same number
  twice. Also: the broker's `/tmp` is a **5 MB tmpfs** and a full one makes
  `--list > file` write zero bytes at exit 0. Both in `LANDMINES.md`.
- **`get_pages_for_rerender` is `include_statuses=['deployed','active']`**, so a
  site re-render leaves `needs_rebuild`/`planned` pages alone (43 across these
  sites). That is a pre-existing backlog, not something this work created.

## 6. The fleet incident that interrupted this — `bugs_open/240`

A correctly-filed item sat 20 minutes undispatched. `kafka-scheduler` had been
**OOMKilled 143 times in 14h**, taking the whole scheduled layer to a ~14% duty
cycle. Cause: `platform/kafka`'s shared transport leaves `MetadataTopics` blank,
which kafka-go documents as "metadata information of all topics in the cluster",
re-fetched on a `rand[0,6s)` background loop — against **25,042 topics**, 24,131
of them orphaned `job.*` topics nothing reliably deletes.

Swept 23,781 topics with owner authorisation. Confirmed causally by dose–response:
24,131 topics → 121Mi and dying; 354 → 15Mi and stable, restarts frozen at 143.

**It stays OPEN and it is on a clock.** `[MEASURED]` topics went **354 → 458 in 37
minutes** (~170/hour) under this lane's own re-render load. Back to five figures
inside a week. `scripts/kafka-orphan-topic-sweep.sh` re-runs the sweep safely
(per-topic liveness test, dry-run default), but candidates 1 and 2 in that file
are the actual fix.

## Cold-start checks

1. `git log --oneline <this-commit>..HEAD -- bugs_open/235_* bugs_open/240_* platform/kafka/dialer.go` — empty = ground unmoved.
2. `go test ./platform/orchestration/actions/ -run 'TestExtractActionInputs_|TestDeployImageAsset_|TestLegacyLogoStep_|TestPurposeFieldBridge_|TestStrategy0DottedPaths_|TestMigration348Shape_'` — 7 expected.
3. Migration 360 still in force: `store_imagery_brand_asset` must show `purpose`
   ABSENT and `purpose_field = input_data.spec.purpose`, while
   `store_hero_asset`/`store_logo_asset` KEEP their statics. Query in `bugs_open/235`.
4. **Topic count** — the 240 clock:
   `kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- bash -c 'bin/kafka-topics.sh --bootstrap-server localhost:9092 --list > /tmp/t.txt 2>/dev/null; grep -c "^job\." /tmp/t.txt; rm -f /tmp/t.txt'`
5. The package's one failing test (`TestValidDocSubjectTypes_Lockstep…`) is the
   064-shape recurrence owned by `idea_uk_vm_site` — pre-existing, not this lane's.
