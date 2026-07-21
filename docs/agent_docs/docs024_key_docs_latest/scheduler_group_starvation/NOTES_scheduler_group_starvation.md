# NOTES — scheduler group starvation (bugs_open/048)

Append-only, newest at the bottom.

## 2026-07-21 — fix, roll, verify (bugfix-048 thread)

- Diagnosis was already filed (048 handoff, by the bugfix-006 thread). Verified it
  against live code + DB before touching anything: `cmd/scheduler/main.go` slot
  taken at (then) line 183 before the pre-query; three early `continue` paths never
  released it; `loadDueTasks` sorts `last_triggered_at ASC NULLS FIRST` so a no-op
  task that never stamps stays head-of-group. All confirmed.
- Live DB confirmed the mechanism: `feasibility-recheck` (head of `maintenance`,
  oldest ts 2026-05-02 04:17) is `fire_message=f`; its pre-query promotes
  `blocked→triaged … RETURNING`, so it returns 0 rows whenever `blocked=0`, which it
  was. `work-item-archiver` has NO pre_query — starved purely by the slot being
  taken. `thunder-reaper` is the same no-rows path in a group of one.
- Fix: (1) moved `inFlight[group]++` down to the commit point; (2) added
  `stampCompleted`, called on the no-op path too. Error paths stay bare `continue`.
  Compiles + vets clean. Committed `dc2e4b61a` BEFORE building (build takes HEAD).
- Build/roll: fresh tag **v1.0.1146** via `make quick-scheduler-update
  IMAGE_TAG=v1.0.1146`. Did NOT edit makefile line 16 (shared, another thread's WIP).

### Misstep I nearly made (the one worth recording)

At first read of the post-roll timestamps, three of the four `maintenance` tasks had
already left May *before* my pod started (`work-item-archiver` 11:45,
`database-cleanup`/`stale-work-item-reaper` 14:47; my pod's first tick was 15:35).
That looked like "someone already fixed it" or "it was never really stuck". It was
neither. **The OLD binary self-heals intermittently:** whenever `blocked>0`,
`feasibility-recheck` returns rows, goes through the CTE-only completion path, stamps
and rotates — briefly un-starving the group — until `blocked` returns to 0 and it
re-pins at the head. So a green-looking timestamp table is NOT proof of the fix; it
can be the old binary in its transient un-stuck phase.

The discriminating test that actually proves the fix: **with `blocked=0`, does
`feasibility-recheck` stamp?** Old binary → no (no-rows `continue`, no stamp). New
binary → yes (`main.go:211` "task ran with nothing to do"). I confirmed `blocked=0`
AND the `main.go:211` log line AND a second advance 600s later. That is the proof;
the timestamp table alone would have been a false pass — a [Verify the failing
branch] shaped trap. Logged the reasoning in the bug file's RESOLVED section so the
next reader doesn't re-walk it.

### Concurrent-session notes

- HEAD moved three times while I worked (my `dc2e4b61a` → `2ba9d3d50` I built from →
  `fe2ba5e52` a "v1.0.1146 sweep" that bumped every kustomization → `2c82cf804`).
  My commit is a durable ancestor of all of them; HEAD's `cmd/scheduler/main.go`
  still contains the fix (checked). The sweep's kustomization bump to v1.0.1146
  matched my deploy's sed — no conflict, no separate commit needed for it.
- The 048 bug file grew ~14 lines under me (another thread's committed edit). My
  Edit reported "modified on disk"; re-read, diffed (only my status line showed),
  appended cleanly. Read-before-write held.
