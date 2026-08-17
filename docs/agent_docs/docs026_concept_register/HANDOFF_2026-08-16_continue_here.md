# HANDOFF — concept register — 2026-08-16

> **SUPERSEDED 2026-08-17 by `HANDOFF_2026-08-17_continue_here.md` — read that instead.** Its figures are re-measured at a current HEAD, it consolidates the still-live items from this doc and from `2026-08-12b`, and it corrects the "the branch is unpushed" line both this doc and 08-12b carry (the remote branch exists, 66 commits behind; only the LOCAL upstream is unconfigured). Nothing in this doc is owed — its three verifier verdicts were read back the same session.

**Cold-start doc for the register lane. This SUPERSEDES
`HANDOFF_2026-08-12b_continue_here.md`.** That doc's picture of the register itself is still
right (all three staleness signals closed, tooled as `DOC-077`/`DOC-078`); what changed since is
**around** the register — a shared-tree incident, its ban, a repair to the landmine delivery
path, and three verifier runs now in flight — plus one new drift finding at HEAD.

Read after this: `RUNNING_NOTES_concept_register.md` § "2026-08-14 / 08-16" (technical log of
everything below, newest at the bottom), then `HANDOFF_2026-08-12b_continue_here.md` for the
register's own state (its "What is actually open now" table still stands), then
`README_where_we_are.md` (owner prose, append-only).

---

## ~~Do these FIRST — three verifier verdicts are owed a read-back~~ DONE, same session

**All three verdicts landed within 2 minutes of dispatch (10:03:44–10:04:05Z) and were read
back: UNVERIFIABLE ×3** — the `verify_unverifiable` branch, reached exactly as predicted because
all three entries' footprints are wholly non-Go (kustomize/Makefile/kubectl/shell/`.md`/`.py`) and
the verifier's `code_symbols` index holds Go only. Each verdict states the entry text is
"internally consistent"; **none objects to anything.** So this is the branch working, not a
refutation (WRONG_CALLS 2026-08-16, first entry). Nothing further owed here. The queries below
stay for the next dispatch — note the correction inside the first one.


Fired 2026-08-16 ~10:03Z, one per entry, via `trigger-landmine-verifier.sh`. **The trigger's
printout is not proof of arrival** (`kcat -P` can drop at exit 0 — LANDMINES) — prove it at the
row, and remember the dispatch queues behind the fleet (publish→start was 29 min once):

```sql
-- The trigger puts the correlation in the ENVELOPE HEADERS; input_data holds only {source, ref}.
-- (My first draft of this used the council-gate key `fix_correlation_id` — wrong trigger, and a
-- jsonb-path scan on this table HUNG for >120s. Key on source + a time bound instead.)
SELECT left(collected_data->'input_data'->>'source', 70) AS entry, current_step, status,
       created_at::timestamp(0), correlation_id
FROM orchestration_states
WHERE created_at > '2026-08-16 09:50+00'
  AND collected_data->'input_data'->>'source' LIKE 'LANDMINES.md#%'
ORDER BY created_at;
-- expected correlation_ids: 4dd05e8a… (stash/manifests), ef045a9a… (silently-inert), 52b70a74… (fewer-files)
```
Then the verdicts themselves:
```sql
SELECT subject_key, created_at::timestamp(0), body FROM doc_notes
WHERE categories ? 'landmine-verification'
  AND (subject_key LIKE 'LANDMINES.md#a-shared-tree-git-stash%'
    OR subject_key LIKE 'LANDMINES.md#appending-a-landmine%'
    OR subject_key LIKE 'LANDMINES.md#git-commit-paths-silently%')
ORDER BY created_at DESC;
```
⚠ Payload shape verified against `scripts/trigger-landmine-verifier.sh` on 08-16 (headers carry
`correlation_id`; `input_data` = `{source, ref}`). And **NEEDS_HUMAN_REVIEW may be a code-index staleness
artefact, not a refutation** (`9f619d938`, 08-15). ⚠ Two of the three entries have footprints
that are wholly non-Go (`git stash`, `kustomization.yaml`, `LANDMINES.md`): expect the
`verify_unverifiable` branch, which reads UNVERIFIABLE — that is the branch working, not a
failure (WRONG_CALLS 2026-08-16, first entry).

**No verdict row after ~45 min → check `orchestration_states` for a FAILED spawn before
re-firing** (spawn→call handshake fails ~half the time fleet-wide; never cancel the failing row
pre-diagnosis). If it never spawned at all, re-fire the same three commands from NOTES §3.

## What changed since 08-12b (all committed)

| commit | what |
|---|---|
| `b3aa8c45c`, `19eb8fdf8` (08-12) | two landmines: a shared-tree `git stash` reverts the production manifests and the tree looks clean; a `·`-separated footprint is SILENTLY INERT |
| `f14a776ed` (08-12) | NOTES 08-12c + 08-12b handoff update: "clean" has two causes; assert the positive fact (`git log -1 -- <path>`) |
| `371317eb6` (08-14) | **`git stash` FORBIDDEN (owner ruling) and mechanically blocked**: `scripts/block-git-stash.py`, PreToolUse hook in `.claude/settings.json`; CLAUDE.md § Git; also carried the owner's 08-12 CLAUDE.md note |
| `f92e0b3ca` (08-14) | **`split_footprints` fixed** (`·` unconditional; commas respect parens; qualifier strip needs a space before `(`); **185 of 482 landmine entries re-keyed** in `doc_notes`; `landmines-sync.py` `existing_sources()` returns subject_keys not counts |
| this commit (08-16) | NOTES, README, WRONG_CALLS row, this handoff; verifier fired ×3 and read back (UNVERIFIABLE ×3, no objections) |

## State in one paragraph

The register is complete, self-consistent, self-monitoring, gated at authoring time
(advisorily), audible, and **all three staleness signals are closed and tooled**. The landmine
delivery path — the mechanism that carries this lane's warnings and everyone else's — is now
keyed correctly for the first time (185 entries were delivering under unmatchable keys). The
tree has a mechanical ban on the one command that erased two days of everyone's uncommitted
work; **no new stash in 680+ commits since**. The count is deliberately not written here (owner
ruling 2026-08-09) — run the drift check.

## The register at HEAD `88897190e` (08-16 11:01) — TWO findings

1. **`PUB-005` — an index row with NO register entry. NEW, and it is the predicted mechanism.**
   The row rode into HEAD as a passenger inside another lane's commit (`88897190e`, the
   `286`/TL-044 commit) while the entry sits complete and DIRTY in `register/public-api.md`
   (+13/−2). Owner is a live session (the `gripper` / tools-api lane). **Not this lane's to
   commit — it closes when they commit `public-api.md`.** Re-run the drift check first: if it is
   still there after a day, contribute a note into their lane, do not commit their file. This is
   the exact state OPP-006 exists to prevent and the 08-12b landmine predicted ("expect the index
   to ride out under another session's commit").
2. **`rebuild-cascade.md`'s stored count — still owed, SIXTH session running.** Its REB-003
   rewrite is still dirty in the tree (+3/−3, restored from the 08-12 stash by this lane, verified
   per file), last commit still `7272d59d4` 07-27. Same-file-blocked; do NOT grow
   `KNOWN_STORED_COUNTS`. **And do not read a clean `git status` as resolution** — assert
   `git log -1 -- <path>` is newer than 07-27 before retiring the line (WRONG_CALLS 2026-08-12,
   this lane).

## Enforcement question — unchanged, still a coin flip

Carried verbatim from 08-12b: **the case for making OPP-006 blocking is not strengthened by a
clean signal.** ~14 entry-adding commits needed for 90% power. Use the per-commit sweep, not a
HEAD snapshot. `PUB-005` above is one data point FOR the leak rate, not against it.

## Landmines specific to this lane (additions since 08-12b)

- **`git stash` is banned and blocked; `git stash list`/`show` are not.** If you find yourself
  wanting one, that is the signal to commit your task narrowly instead. The 08-12 stash
  (`stash@{0}`, base `1ee940968`) is still in place and still holds work — extract by path only.
- **A `NEEDS_VERIFICATION` line from `landmines-sync.py --apply` is an UNSENT dispatch.** Two of
  this lane's entries sat armed-and-unsent from 08-12 to 08-16. Run
  `./scripts/landmines-verify-dispatch.sh` (sync + dispatch) rather than `--apply` alone, or
  `trigger-landmine-verifier.sh` per slug afterwards. CLAUDE.md was corrected on this 08-15.
- **A landmine footprint must be tested for MATCHING, not for syncing.** The check is in the
  `SILENTLY INERT` entry: parse it and assert a LIST of short grep-able strings. Now that `·`
  splits, the house convention is legal — but a glob still never matches a real path.
- **`landmines_lib.py` has a self-test now**: `python3 scripts/landmines_lib.py` (8 cases). Run it
  before touching `split_footprints`; my own first fix failed it.

## Owed / open

- ~~Verifier verdicts ×3~~ — DONE: UNVERIFIABLE ×3, no objections (top of this doc).
- `PUB-005` — the other lane's to close; watch only.
- `rebuild-cascade.md` — still owed, still blocked.
- The **58 → 0** collapsed footprints: DONE. The `SILENTLY INERT` landmine's dated update records
  it. Nothing further here unless a verdict objects.
- **The branch is still unpushed** (`## 087_towards_multiple_domains`, no upstream shown). The
  register CronJob reads the pushed branch. Not this lane's call.
- Stray `register/model-infrastructure.md.tmp_check` — still untracked, still not ours.

## Things deliberately NOT done

- **No SUMMARY.** Five-headings test: the register's own state is unchanged since 08-12's summary;
  the stash incident and the delivery-path repair are real work but they are not a change in
  "where the register is". When the verdicts land and the enforcement question moves, that is
  the next inflection.
- **`PUB-005` not committed** — another live lane's file.
- **`rebuild-cascade.md` not touched.**
- **Nothing pushed.**
