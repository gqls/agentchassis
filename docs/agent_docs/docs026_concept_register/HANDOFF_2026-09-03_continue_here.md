# HANDOFF — concept register — 2026-09-03

**Cold-start doc for the register lane. This SUPERSEDES `HANDOFF_2026-08-17_continue_here.md`.**
Everything below was re-measured at HEAD `566d59db3` on 2026-09-03, after the owner's fresh
chassis roll. **Every figure carries its command — re-run before repeating any of them
outward; this tree moves ~250–400 commits a day** (it moved ~1,200 while the last sitting's
work was being written up).

Read after this, only as needed: `RUNNING_NOTES_concept_register.md` (technical log, newest at
the bottom — the 08-25, 08-25-later and 08-31 sections cover the last three sittings) and
`README_where_we_are.md` (owner prose, append-only). You should not need the earlier handoffs;
their live content is consolidated here.

---

## Start here — the lane is QUIET, nothing is mid-flight, and there is one obvious next job

No dispatch in the air, no verdict awaited, no half-applied change. **The eight-session
`rebuild-cascade.md` blocker is CLOSED.** In priority order:

1. **⚠ SETTLE THE ROLL-PENDING BACKLOG — this is new, it is time-boxed, and it is the best
   work on the list.** The owner rolled the fleet on 2026-09-03 (`v1.0.1356`). `[MEASURED
   2026-09-03]` **110 register entries still say "inert until the next roll"** — every one of
   them may now be live and none of them knows it. **81 name a commit and can be settled
   mechanically**, one command each:
   `git merge-base --is-ancestor <the entry's sha> <the service's stamp>` → in or out, no
   judgement. **29 name nothing and cannot be settled at all** — that is the population
   `OPP-011` (below) now prevents growing, but it does not shrink it.
   This is exactly what `BLD-019` did on 2026-08-10 when the stamp first landed ("the first
   thing it did was retire a guess"), and it has not been done since. Reproduce the list:
   the census in `RUNNING_NOTES` 2026-08-31, or re-derive with the parse in
   `scripts/report-register-version-lag.py`.
   ⚠ **Correct each entry visibly** (strike-through + date), never silently — and note the
   register convention that a struck-through claim is how withdrawal is recorded, which is
   why `OPP-011` strips `~~…~~` before testing.
2. **Item 2, still the lane's real open question: are `DOC-077`/`DOC-078` ACTIONABLE?**
   Unchanged and unanswered — take one entry from
   `./scripts/report-register-version-lag.py --worklist`, repair its expired evidence by hand,
   and record how long it took and whether the report pointed at the right thing. A month of
   "nobody acted on it" is a finding, but only if somebody tried once. **Note item 1 is a
   cheaper version of the same experiment** and may answer it as a side effect.
3. **The enforcement question (`OPP-006` blocking) — do NOT touch without reading its section
   below.** Still Delivery PROVEN / behaviour OPEN.
4. **The 4 dead citations** — route them to their owning lanes; do not guess the paths.

## State in one paragraph

The register is complete, self-consistent, self-monitoring, gated at authoring time
(advisorily, now by two checks rather than one), audible, and all three staleness signals are
closed and tooled. The lane's four-handoff item 1 shipped on 2026-08-31 and is council
APPROVED. The citation report's sharpest category had a false-positive mode; that is fixed and
the category is back to its true four. The one blocker that had survived eight sessions is
gone. **The count of entries is deliberately not written here** (owner ruling 2026-08-09): it
is derived, so run the drift check.

## Verified at HEAD `566d59db3`, 2026-09-03 — with the command for each

| what | reading | command |
|---|---|---|
| register self-consistency | **2059 entries, 2059 index rows — agree exactly** | `./scripts/test-concept-register-drift-local.py` |
| register drift findings | **1** — `WII-035` listed by two index rows (work-item-integrity lane's, the `OPP-006` arm-2 collision shape). **Not this lane's.** | same |
| landmine rows ↔ file | **930 entries both sides, 0 key mismatches**; 5,163 rows | `./scripts/landmines-keys-check.py` |
| citation health (`DOC-078`) | **10,443 citations / 1,985 entries; 7,858 resolve (75%); 4 NEVER-REPO-PATH** | `./scripts/report-register-citation-rot.py` |
| roll-pending backlog | **136 roll-conditional · 26 withdrawn · 110 live · 29 undated (26%)** | census in `RUNNING_NOTES` 2026-08-31 |
| live fleet | **`v1.0.1356`** on 20 deployments; makefile `IMAGE_TAG` is `v1.0.1357` (next build bumped, not yet rolled) | `kubectl -n ai-persona-system get deploy -o jsonpath=…` |
| the stash ban | still holding; `stash@{0}` is still the 2026-08-12 one | `git log -g --pretty='%gd %ad' --date=iso refs/stash \| head -1` |

**The 75% citation ratio has now held across FOUR measurements** (08-12, 08-17, 08-25, 09-03)
while the corpus grew 8,279 → 10,443 citations. That is the useful reading and it is stable:
the citations are abbreviated, not rotting.

## What changed since 2026-08-17

| commit | what |
|---|---|
| `a9665268f` | **`DOC-078`'s sharpest category had a false-positive mode.** `clean()` stripped line refs with `:L?\d+([-,]L?\d+)*$` — fine for `:151,227`, wrong for the repeat-colon form `:65,:283`, which it half-stripped into a path git never held, landing it in `NEVER-REPO-PATH` ("no file, ever"). One character class fixed it. Guarded by a self-test case **proven non-vacuous by mutation** (revert the regex on a copy, keep the case, watch it fail with exactly `NEVER-REPO-PATH`). |
| `d4e462950` | **`rebuild-cascade.md`'s stored count retired — the EIGHT-session blocker, closed.** Owner ruled 2026-08-26 to take the same-file passenger deliberately and declare it. It carries another lane's REB-003 rewrite (`bugs_open/182`), which had sat uncommitted since being restored from `stash@{0}` after the 08-12 stash incident. |
| `cdc70e2dd` | **Notified the `loancalculator_couk` lane** in their cold-start doc that their work is now in HEAD under someone else's commit message — where their next session actually reads, because filing in the right directory is authoring, not delivery. |
| `1efc84362`, `7006eb2d8`, `7dbe5b8fd`, `80c6a9a83` | **Item 1 shipped: `check_register_roll_claim_without_commit`, registered as `OPP-011`.** Council **APPROVED round 1** (`Council-Reviewed: 37b0bec4-f503-4b9a-8fc4-688ba29aa2bc`), 2 advisory objections, both answered. |
| LANDMINES | new entry: *"A `NEVER-REPO-PATH` citation naming a file you can `ls` is an UNCOMMITTED file, not a dead citation"* — synced (7 rows), dispatched, verdict UNVERIFIABLE (correct for non-Go footprints). |
| another lane | **the `landmines-sync.py` transport defect is FIXED** (`02c740616` retry on mid-stream EOF, `f3cfbbf78` the mislabelled print). `delta vs the file: 0/0/0/0` verified 2026-08-26. Nothing owed. |

## ⚠ THE ENFORCEMENT QUESTION — read before touching `OPP-006`'s blocking behaviour

Unchanged in substance: **a clean signal is not evidence.** ~14 entry-adding commits for 90%
power at the measured 16% leak rate, ~18 for 95%. Use the **per-commit sweep**, never a HEAD
snapshot — a snapshot cannot see a leak created at 10am and repaired by 2pm.

**Evidence still points both ways, and this lane widened the range rather than settling it.**
Two 08-16 leaks self-healed inside ten minutes; one measured 2026-08-20 (`SEO-005`) took
**~50 minutes**. The argument NOT in the older handoff, and worth weighing: council seats read
register status lines as ground truth, so a leak window is a window in which a seat can cite a
phantom concept. Against that, satisfying `OPP-006` costs one line — the usual "you cannot hold
work on a shared tree" objection does **not** apply here. The owner has been offered a
per-commit sweep to settle it and has not yet asked for one.

## Landmines specific to this lane

- **The drift trio read a REF, never your worktree** (`DOC-078` is the exception). An entry you
  have not committed is invisible to the harness — and a `NEVER-REPO-PATH` naming a file that
  exists on disk is an **uncommitted file, not rot**. Check `git status <cited path>` before
  believing that category. Full entry in `LANDMINES.md`.
- **A row-without-entry drift finding younger than ~10 minutes is a lane mid-commit, not a
  leak** — but 2026-08-20 measured one at ~50 minutes, so age alone is weaker evidence than the
  older handoff implied. Re-run the check and look at the tree before filing anything.
- **`landmines-sync.py --check` says "in sync" WITHOUT checking key identity.** The real
  assertion is `./scripts/landmines-keys-check.py`.
- **After appending a landmine run `./scripts/landmines-verify-dispatch.sh`** (sync AND
  dispatch), never `--apply` alone — it consumes the "new entry" status. ⚠ **And a `psql
  failed:` from that script may have ALREADY COMMITTED its write**: on 2026-08-25 four of five
  attempts printed transport errors, the rows were already in, and the status had been consumed
  with no dispatch. **Query for your rows before retrying.** (The transport itself is fixed
  now; the "failure that already succeeded" shape is not specific to it.)
- **A landmine footprint must be tested for MATCHING, not for syncing.** Parse the entry and
  assert a LIST of short grep-able strings — a first draft of this lane's own entry carried a
  103-character prose footprint and the parse caught it.
- **`LANDMINES.md` entry headings are `###`, not `##`** — and `grep '^## '` will confidently
  show you the non-conforming minority as if it were the convention (717 `###` vs 148 `##`).
  A wrong level still parses and still costs DELIVERY; `landmines-sync.py` names the offenders
  in a block titled *"N warning(s) that cost DELIVERY"*.
- **A `verify_unverifiable` / UNVERIFIABLE verdict is the branch WORKING**, not a refutation —
  the verifier's index holds Go only, so any wholly non-Go footprint lands there.
- **Four commands count the index and all four are correct. Do not write any of them down.**
- **A daily report's count is not a RATE** — the watcher only sees what is still broken at 06:50.
- **`git diff | grep '^-[^-]'` cannot see a deleted markdown BULLET** (`- **what:**` reads as
  `--` in a diff). Gate on `git diff --numstat`; to read what went, `grep '^-' | grep -v '^---'`.
  This lane walked into it on 2026-08-26 having been warned at session start.
- **The CronJob reads the PUSHED branch** and `REGISTER_REF` is hand-pinned in
  `deployments/kustomize/services/concept-register-drift-check/base/cronjob.yaml`.
  **The owner pushes, deliberately and in quiet periods, because a push reboots containers** —
  this lane does not push. Check the remote directly:
  `git ls-remote --heads origin 087_towards_multiple_domains`, then `git rev-list --count <sha>..HEAD`.
- **Two ConfigMaps exist** for the CronJob (kustomize does not prune) — only the mounted one is live.

## Owed / blocked

- **Nothing.** No dispatch in flight, no verdict awaited, no half-applied change, and the
  `rebuild-cascade.md` blocker that stood for eight sessions is closed.
- Stray `register/model-infrastructure.md.tmp_check` — still untracked, still another lane's.
- `WII-035`'s duplicate index row is the work-item-integrity lane's, not ours.

## Things deliberately NOT done

- **No SUMMARY** (owner agreed 2026-08-26). Five-headings test: the register's own state is
  materially what `SUMMARY_where_we_are_2026-08-12.md` says. **The next inflection is uptake** —
  items 1 and 2 above are what would produce one.
- **The 881 repairable citations and the 4 dead ones**: not rewritten, by design. An automated
  rewrite across 111 files is a change no reviewer can check, and a guessed path that lands on
  the wrong file is worse than a dead one — a dead link fails loudly, a wrong link fails silently.
- **The 29 undated roll-pending entries were not back-filled** by looking up their commits.
  `OPP-011` stops the population growing; shrinking it is item 1 and is a reading job, not a
  scripted one.
- **The `_RELOCK` migration-suffix warning** printed by `scripts/council-scope.sh` is
  `bugs_open/314`'s, not this lane's.
- **No push.** Owner's, by arrangement.
