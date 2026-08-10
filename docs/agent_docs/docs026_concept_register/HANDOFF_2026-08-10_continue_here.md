# HANDOFF — concept register — 2026-08-10

**Cold-start doc for the register lane.** Read this, then
`SUMMARY_where_we_are_2026-08-09.md` (the read-aloud milestone),
`README_where_we_are.md` (owner-facing prose, append-only) and
`RUNNING_NOTES_concept_register.md` (technical log, newest at the bottom).
Written because the originating chat grew long, not because the work stalled.

---

## State in one paragraph

The concept register is **complete, self-consistent and self-monitoring**. Entries
and index rows agree; no id is used twice; stored counts have been retired
everywhere but one file. A daily CronJob (`concept-register-drift-check`, DOC-074)
reads the register from GitHub and writes a verdict to `doc_notes` — it is **live,
firing unattended since 2026-08-05, and it has caught real drift on four separate
days**. What is *not* established: whether the ~1,800 entries are still **accurate**.
Verification ran once, in July. Coverage answers "is it here?", drift answers "does
it agree with itself?", and nothing yet answers "is it still true?"

## How to check the register in 15 seconds

```bash
./scripts/test-concept-register-drift-local.py            # the live check's own logic, against HEAD
./scripts/test-concept-register-drift-local.py --self-test  # + the historical control and two mutations
```
Or read what the cluster saw this morning:
```sql
SELECT created_at, body FROM doc_notes
 WHERE subject_key = 'concept-register-drift'
 ORDER BY created_at DESC LIMIT 1;
```
⚠ The CronJob reads the **pushed branch**; the harness reads whatever ref you give
it. On this tree those differ routinely — that is not a bug, but a "clean" verdict
is never a statement about your working tree.

## What exists, and where

| thing | path | state |
|---|---|---|
| the register | `docs/agent_docs/docs026_concept_register/register/` | 109 category files + `000_concept_index.md` |
| the watcher | `deployments/kustomize/services/concept-register-drift-check/` | LIVE, daily 06:50 UTC |
| its local twin | `scripts/test-concept-register-drift-local.py` | same functions, git instead of GitHub |
| coverage check | `docs026_concept_register/102_CHECK_register_coverage.py` | on the commit path (OPP-004) |
| register entry | `register/documentation-system.md` → **DOC-074** | full design + 3 landmines |

Deploy/trigger: `make deploy-concept-register-drift-check ENVIRONMENT=production
REGION=uk001`, then `make concept-register-drift-check-now` and
`…-logs`. **`make release` does NOT cover CronJobs** — its service list is
hardcoded, which is why a fleet release leaves this untouched.

## What was done (2026-08-04 → 08-10)

1. **Brought the register up to date.** Found 34 concepts with an entry and **no
   index row** (all of `CLM-001…012` among them); backfilled. Found the coverage
   ratchet ignored every annotated line, so 12 of 17 "new" subsystems each run were
   already-settled decisions; fixed in `ratchet_name()`.
2. **Deleted 1,339 superseded duplicate documents** (441 documents existing as
   1,973 numbered copies). All recoverable from git; 43 register `sources:` lines
   now resolve only through git, which is recorded in LANDMINES.
3. **Built and deployed the watcher** at the owner's request — in the framework,
   not the CLI. Copies `bugs-open-staleness-sweep` (DOC-071) wholesale.
4. **Retired every stored count** (owner ruling 08-09) after measuring that four
   commands count the index with four individually-correct answers, and that 32 of
   109 per-file counts were already wrong (90 concepts of drift).
5. **Fixed what the watcher found**: `SCH-024`, `BIZ-031`, `WFA-012` (missing rows)
   and the `LNK-031` id collision (renumbered `LNK-032`).

## The measurement that should drive what happens next

**The missing-row defect recurs at roughly one every day and a half.** Not
estimated — observed, by something counting:

| date | concept | how it was found |
|---|---|---|
| 08-04 | 34 concepts at once | manual comparison (the founding discovery) |
| 08-08 | `SCH-024` | local drift check |
| 08-10 | `BIZ-031`, `WFA-012` | local drift check |

Plus a **headline mismatch reported on three consecutive days and corrected by
nobody** — the report was right and unread. That is the single most important
observation in this lane: *a mechanism that writes into a table nobody opens fails
the same way the convention it replaced did.*

## Next, in the order I would do it

1. **Build the pre-commit gate for the missing-row class.** The evidence is now in:
   ~1 per 1.5 days, four separate lanes, none of them careless. A `pattern-check.py`
   rule that fires when a commit adds a `### <ID> —` heading without adding its
   `| <ID> |` index row in the same commit. Same shape as OPP-003/OPP-004; the
   check is ~20 lines and the corpus to test it against is the last week of
   commits. **This removes the drip at the point of authoring**, which the daily
   report cannot do.
2. **Clear the one owed stored count.** `register/rebuild-cascade.md` still states
   `7 concepts`. It was skipped because another session has had that file dirty in
   the shared tree since 08-04, and a pathspec commit takes a same-file passenger.
   If it is clean when you read this, retire the line and delete
   `rebuild-cascade.md` from `KNOWN_STORED_COUNTS` in the local harness. **Do not
   grow that set to silence findings.**
3. **Then staleness — the real open flank, and a design question, not a chore.**
   Nothing checks whether an entry is still TRUE. A register status is already a
   known landmine: a snapshot that outlives its truth, quoted by council seats as
   ground truth. Building blocks exist — `covers-through` stamps, `landmine-verifier`
   (DOC-069), the bugs-open staleness sweep (DOC-071). Worth its own session.

## Landmines specific to this lane

- **The watcher reads the PUSHED branch.** A clean verdict says nothing about your
  tree. Reproduce either side: `./scripts/test-concept-register-drift-local.py <ref>`.
- **`REGISTER_REF` is hand-pinned in the manifest**, and is the *second* such ref in
  the estate (`SWEEP_REF` is the first). When the working branch moves, both need
  bumping. A stale-but-resolving ref is the worst case: every finding becomes
  unfalsifiable **and every "clean" run becomes meaningless.**
- **`git fetch` with no refspec sets `FETCH_HEAD` to whatever branch it wrote
  last.** It resolves, errors nothing, and answers about the wrong branch — it told
  me the retirement was unpushed when it was pushed. Use `git ls-remote origin
  refs/heads/<branch>` and check ancestry against that sha.
- **The index's frozen log quotes the old headlines verbatim**, so both count
  searches in `check.py` are head-bounded. A whole-file search would report a
  finding every run for ever — a watcher crying wolf about its own archive.
- **Four commands count the index and all four are correct.** Only
  `grep -cE '^\| [A-Z]{2,4}-[0-9]{3} \|'` is the row count. Do not write any of them
  down anywhere.
- **Two ConfigMaps exist for this service** (kustomize does not prune). Only the one
  the CronJob mounts is live: `kubectl get cronjob concept-register-drift-check -o
  jsonpath='{…volumes[0].configMap.name}'`.

## Things I deliberately did NOT do

- **No existing entry's status was re-verified.** Not one. The register's honest
  claim is complete-and-consistent, never current.
- **`WFA-012`'s row does not claim live**, though a chassis rolled on 08-10 and its
  entry says "inert until the next image". Its change is control flow with **no new
  string literal**, so no pod-grep marker exists — `ExtractNestedField` predates it
  and greps 8 times either way, a positive control that proves nothing (DOC-073's
  case). Its lane owns that proof.
- **No `--update-ratchet` run.** It now preserves annotations, but rewriting the
  ratchet is a deliberate act and the report is quiet, so there was no reason.
