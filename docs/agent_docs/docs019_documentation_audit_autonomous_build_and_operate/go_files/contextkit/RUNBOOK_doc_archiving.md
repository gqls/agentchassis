# Runbook — documentation archiving subproject

**Objective:** reduce the docs tree to its current, usable documents so the
analyser's index isn't diluted by dozens of stale versions and duplicate copies.
Every step is REPORT-FIRST and REVERSIBLE; nothing is deleted (only moved to
`_archive/`, with a manifest). You run every command.

**Where the noise actually is** (measured 2026-06-13 from
`main_docs_directory_tree.txt`): 2,729 files collapse to 1,917 subjects; **1,734
are singletons** (one file — untouched by everything here); the noise is
concentrated in **18 fat clusters of 10+ versions each** (~489 files), almost
all under `docs024_key_docs_latest/` (`running_notes`, `traffic_probe_*`,
finetuning `phase5`, `016_debugging_guide`). The strategy targets those, not a
blanket pass.

**Two kinds of redundancy, two tools** (don't conflate):
- identical/near **copies** (same content, `(N)` download dups) → `dedup`
- older **versions** (different content, successive edits) → `thin_versions`

**Order matters:** dedup first (clears exact copies so version-thinning doesn't
mistake a download-dup for a rival version), then thin_versions, then the
editorial moves, then re-index.

---

## Pre-flight

```bash
# Tools live in the contextkit module. From the module root:
go build ./cmd/dedup ./cmd/thin_versions    # both compile
# Take a safety snapshot of the docs tree FIRST (the moves are reversible via
# manifest, but a snapshot is the cheap belt-and-braces):
DOCS=docs/agent_docs            # <-- set to your real docs root
tar czf /tmp/docs-snapshot-$(date +%F).tgz "$DOCS"
```

A caution carried from the dedup `-move` bug: these tools were behaviour-tested,
but they MOVE FILES. Run each in its default REPORT mode first and read the
output before adding `-move`. Report-only is the default precisely so a
surprise is "it did nothing", never "it moved the wrong thing".

---

## Step 1 — exact + near-duplicate COPIES (dedup)

```bash
# 1a. REPORT exact duplicates (SHA-256 — zero false positives):
go run ./cmd/dedup "$DOCS" -ext .md,.go,.json,.sql

# 1b. REPORT near-duplicates too (heuristic — shingled-token similarity):
go run ./cmd/dedup "$DOCS" -ext .md -near -threshold 0.9
#     Read the near groups. They are NOT exact; eyeball that the "archive"
#     picks really are redundant before trusting -move -near.

# 1c. ACT — exact first (safe), review, then near if the report looked right:
go run ./cmd/dedup "$DOCS" -ext .md,.go,.json,.sql -move
go run ./cmd/dedup "$DOCS" -ext .md -near -threshold 0.9 -move    # only if 1b looked right
```
- Canonical kept per group: non-archive > non-`(N)` > shallowest > shortest >
  newest. So the real file is kept and the graveyard/download copy archived.
- Output: copies moved to `$DOCS/_archive/<original-path>`; `dedup-manifest.tsv`
  records every move with undo guidance.
- **Undo:** for each manifest row, `mv` `moved_to` back to `moved_from`.

---

## Step 2 — older VERSIONS of each subject (thin_versions)

```bash
# 2a. REPORT the fat clusters and the keep/archive split (default keep 5, only
#     groups with >= 10 files):
go run ./cmd/thin_versions "$DOCS"

# 2b. Also surface number-bumped-same-subject pairs (the 004_/005_ mixup case)
#     for your review — NOT auto-merged:
go run ./cmd/thin_versions "$DOCS" -report-renames

# 2c. Tune if you want a different keep count or to reach smaller clusters:
go run ./cmd/thin_versions "$DOCS" -keep 10 -min-group 5     # report

# 2d. ACT:
go run ./cmd/thin_versions "$DOCS" -keep 5 -move
```
- Recency rank within a subject: **version (`_vX_Y`) > `(N)` bracket > mtime**.
  Version beats mtime deliberately — a `v2_31(1)` with an older timestamp than
  `v2_30` still ranks above it. This is the "updated out of turn" case you
  flagged: a stale-DATED file that is a later VERSION ranks correctly.
- Singletons and already-archived files are never touched.
- Output: older versions moved to `$DOCS/_archive/<original-path>`;
  `thin-manifest.tsv` records every move.

**The one judgement to apply (read 2a before 2d):** the tool keeps the NEWEST N,
which is a recency heuristic, not a quality check. If a cluster's best version is
genuinely an older one (rare for running-notes; possible for a doc someone
rewrote and regretted), it would be archived. The move is reversible and
manifested, so this is recoverable — but it is why you read the report first.

---

## Step 3 — re-home the surviving CURRENT docs (editorial, human-led)

The duplicates and old versions are gone; what remains is the current set, still
where it always lived. To give the analyser a clean structure (the
`engines/` + `runbooks/` tree in `engines_tree_proposal.md`):

```bash
# 3a. Deterministic part + the editorial proposal (REPORT):
./cmd/dedup/stage_docs019_migration.sh "$DOCS"
#     - auto-moves obvious archive dirs (go_files_old/, thin_slice_run/,
#       working/) into _archive/ on --apply;
#     - writes PROPOSED_MOVES.tsv mapping each loose current doc to a likely
#       engines/ or runbooks/ home, with (N) files flagged "leave for dedup".

# 3b. Apply the deterministic archive-dir moves:
./cmd/dedup/stage_docs019_migration.sh "$DOCS" --apply

# 3c. EDIT PROPOSED_MOVES.tsv — set the ACTION column (move|archive|skip|keep)
#     per row. THIS IS YOURS: which file is canonical / what each is about is a
#     judgement a script can't make. Then apply with the one-liners the script
#     prints (git mv the "move" rows; archive the "archive" rows).
```
- `engines/tool-docs/` currently mixes module files with docs — untangle it by
  hand; no heuristic separates those safely.
- Per the documentation discipline: **classify, don't merge.** If you consolidate
  overlapping docs, carry passages across by hand with the LLM as an assistant
  that finds and cites unique content — never a generative merge that can drop a
  caveat silently.

---

## Step 4 — re-point and re-index

```bash
# 4a. Fix internal links to any moved/renamed files; update the area README map.

# 4b. Rebuild the analyser index over the CLEAN tree. With everything stale now
#     under _archive/, the exclude collapses to ONE entry:
go run ./cmd/analyser "$DOCS" -exclude _archive/ > chassis.json

# 4c. Verify zero stale/duplicate paths survived into the index:
python3 -c "import json,re; ps=[f['path'] for f in json.load(open('chassis.json'))['files']]; \
  bad=[p for p in ps if re.search(r'\(\d+\)\.go|_archive/|go_files_old/|docubundle/', p)]; \
  print('stale paths in index:', len(bad)); [print('  ', p) for p in bad[:10]]"
#     Expect 0. Non-zero → an exclude missed a path; add it and re-run 4b.
```

---

## Verification gates (each step is checkable before the next)

| After | Check | Expect |
|---|---|---|
| 1c dedup | `wc -l $DOCS/dedup-manifest.tsv`; spot-check a kept canonical opens fine | copies in `_archive/`, canonicals in place |
| 2d thin | `ls` a thinned cluster dir | newest N remain; tail in `_archive/` |
| 3b stage | `ls $DOCS/_archive/` | `go_files_old/` etc. now archived |
| 3c moves | `git status` | only intended moves staged |
| 4c index | the python one-liner | `stale paths in index: 0` |

## Reversibility (the whole subproject is undoable)

- Every move tool writes a manifest (`dedup-manifest.tsv`, `thin-manifest.tsv`)
  with `moved_from` → `moved_to`; reverse any move with `mv moved_to moved_from`.
- Nothing is deleted; `_archive/` holds everything set aside.
- The pre-flight tarball is the full fallback.
- `git` (if the tree is tracked) is the final safety net — review `git status`
  before committing, and the moves are a single revert away.

## What this subproject does NOT do

- It does not MERGE documents (classify-don't-merge — Step 3 note).
- It does not judge document CONTENT currency — that is the doc-drift classifier
  (`DESIGN_doc_drift_classifier.md`), a separate, later, T1+T2-first tool.
- It does not touch the 1,734 singleton subjects — only the fat clusters and the
  duplicate copies.
EOF
echo created; wc -l RUNBOOK_doc_archiving.md