# 402 — `landmines-sync.py` delta logic wants to rewrite the WHOLE corpus, so every sync now dies in the kubectl exec stream

Filed 2026-08-25 ~21:00Z by the gripper-dossier ("AI page 3") lane, which hit it while
delivering one new LANDMINES entry. **Symptom report with evidence — NO root cause is
asserted here** (per the 2026-07-31 owner ruling this file makes no structural claim;
the discriminating first checks are listed for whoever picks it up).

## Symptom

`./scripts/landmines-verify-dispatch.sh` (and `landmines-sync.py --apply`, and
`--check`) fail identically, 3/3 runs on 2026-08-25 20:55–20:57Z:

```
doc_notes: 4664 owned row(s) across 846 entr(ies)
  to insert/refresh: 847   orphaned (entry retitled/removed): 0
psql failed:
error: error reading from error stream: read message: unexpected EOF
```

Two distinct facts in that output:

1. **The delta logic classifies ALL 847 entries as needing insert/refresh** (846
   pre-existing + the 1 genuinely new one). The 2026-08-12 delta rework exists
   precisely so a routine sync sends only what moved; "everything moved" on a day
   when one entry was appended means the comparison is failing wholesale, not that
   the corpus changed.

   > **CORRECTED 2026-08-26 (closing session):** the "847" was `len(want)` — the
   > file's TOTAL entry count, printed by that line on every run, healthy or not,
   > before any comparison had executed. The delta logic was fine; the real delta
   > that night was ~1 new entry. See the closure section below, and WRONG_CALLS.md
   > 2026-08-26 ("read the print before theorising about the number").
2. **The full-corpus payload then dies in the `kubectl exec` transport** — the exact
   `unexpected EOF` the script's own comment (line ~256) records from 2026-08-12,
   which the delta rework was built to avoid. By that comment's own count this is
   the FOURTH time the sync has broken purely because the corpus grew.

## What was ruled out tonight

- **Not cluster access**: `kubectl exec -i postgres-clients-0 -- psql -Atc 'SELECT 1'`
  works in the same minute. The pod is healthy (`Running`, 30d).
- **Not the new entry's format**: it uses `###` + the standard bullet set; the
  read-back stage lists it with 3 footprints, correctly parsed. (164 PRE-EXISTING
  `##`-headed entries are flagged as delivery-costing warnings — dated long before
  today.)
- **Not a recent edit to the script**: `git log --since=2026-08-20 --
  scripts/landmines-sync.py` is empty.
- The corpus READ leg works (it printed the 4664-row/846-entry census); the failing
  psql is the one after the delta computation.

## Impact

- **Landmine DELIVERY to `doc_notes` is broken estate-wide** — no lane can currently
  sync a new entry, so council seats and agents read a corpus whose freshness is
  unknown. When it last synced successfully is NOT established (worth establishing:
  `SELECT max(created_at) FROM doc_notes WHERE categories ? 'landmine'`).
- **The verifier cannot be armed for new entries** (`landmines-verify-dispatch.sh`
  dies at its sync stage, before arming).
- The FILE stays the system of record (owner ruling D10), so nothing is lost — but
  the 2026-08-25 compose-drift entry (gripper lane) is undelivered; re-run the
  dispatch for it when this is fixed.

## Discriminating first checks for the fixing thread

1. **Why 847-of-847?** Pick ONE old entry; print the body the script WOULD write vs
   the body doc_notes HOLDS (`SELECT content FROM doc_notes WHERE source = '<src>'
   LIMIT 1`). A systematic difference (prefix, escaping, category shape, a field
   added to the composed body) means every row compares stale and the "delta" is a
   full rewrite by construction — the EOF is then downstream damage, not the defect.
2. **If the bodies match**, the comparison code itself regressed — diff what the
   compare reads against what the writer writes (the 08-14 refootprint fix touched
   exactly this seam).
3. Only then look at transport: batching the apply into N smaller psql calls fixes
   the EOF regardless, and the script's history says the corpus will keep growing —
   but batching alone leaves check-vs-apply asymmetry if (1) is the defect.

## Workaround status

None taken. Hand-writing `doc_notes` landmine rows is forbidden (CLAUDE.md: "append
to the file, never hand-write a landmine row into doc_notes"), and `--full` sends the
same payload that dies.

---

*Renumbered 401 → 402 within the hour of filing (2026-08-25 ~21:05Z): another
session was concurrently filing its own 401 (`…discovery_watchdogs_driver_alarm…`,
untracked at the time). Commit messages `4f44a4d0c`/`c65d44350` say "401" and are
immutable (forward-only) — resolve by SLUG, as always.*

---

## CLOSED 2026-08-26 — diagnosed and fixed (session "landmines.sync")

The symptom's two "distinct facts" resolved OPPOSITE ways.

**Fact 1 REFUTED — the delta logic was never broken.** The line
`to insert/refresh: 847` printed `len(want)`, the total entry count parsed from the
FILE, on every run since the 2026-08-12 delta rework — before any comparison had
executed. (The real delta lists — `new`, `changed`, `refootprinted` — were computed
further down, and `refootprinted` only inside the `--apply` branch, so a dry run
never showed it at all.) A healthy sync printed the same shape.

**Fact 2 CONFIRMED, root cause on the READ leg.** The failing psql call was the one
directly after the census line: the body read-back, which shipped every entry's
prose through `kubectl exec` to answer a per-entry yes/no question
[MEASURED 2026-08-26: bodies 2,962,987 B, subject_keys 610,076 B]. That stream dies
mid-transfer roughly every other call at ~3MB — the fourth corpus-growth failure in
this script's history, per its own docstring. The full-corpus *write* this file's
title feared never happened; the apply never got that far.

**Fixes, both live (scripts are live at commit):**

- `02c740616` (2026-08-25 21:08 BST, a **concurrent session**, minutes after this
  file was filed): 3-attempt retry on the transport-EOF signature in `run_psql`.
  Delivery resumed the same evening — the synced corpus grew 846 → 857 entries
  between filing and closure.
- The commit that moves this file to `bugs_closed/` carries the structural half:
  **the read-back no longer scales with the corpus.** One query returns per-entry
  `(row count, md5 of the footprint set, md5 of the body)` — ~130 B/entry, ~110KB
  total, instead of 3.6MB — compared against Python-side md5. The footprint
  aggregation orders `COLLATE "C"` to match Python's code-point sort (UTF-8 byte
  order preserves it); verified against the old full-payload comparison on all 856
  live entries, zero disagreements, BEFORE adoption. Also fixed in the same commit:
  the delta print now prints the DELTA (fact 1's misdiagnosis surface); `--check`
  now fails on changed/refootprinted entries too (it silently passed body edits
  before); and the "N warning(s) that cost DELIVERY" banner no longer counts the
  164 `##`-heading nags (those entries parse and deliver — proven by the
  `##`-headed compose-`.env` entry's 4 delivered rows and two verifier verdicts).

**Apply-path verification:** deleted one owned entry's rows
(`…the-repo-copy-of-a-box-deployed-config…`, 3 rows) and re-ran `--apply`: detected
as `1 new`, 8,517 bytes on the wire, row count restored to 4,723, `--check` exit 0
after (hash-identical to the file), `NEEDS_VERIFICATION:` line printed for the
wrapper.

**The impact section's open items, resolved:**

- *"When it last synced successfully is NOT established"* — established: owned rows
  created through 2026-08-26 morning, and a successful verifier dispatch the
  evening of filing (verdicts below).
- *"the 2026-08-25 compose-drift entry (gripper lane) is undelivered; re-run the
  dispatch for it"* — already done before closure: source
  `LANDMINES.md#the-repo-copy-of-a-box-deployed-config-file-drifts-behind-live-edits-made-on-the`
  has its rows AND two landmine-verifier verdicts dated 2026-08-25 20:05:04Z and
  20:08:20Z (both UNVERIFIABLE — which is a verdict: the dispatch ran). Nothing to
  re-fire.

**Residual, deliberately not fixed here:** **18** entries [COUNTED 2026-08-26] are
genuinely undelivered — they have no `footprint:` line, so the parser skips them
(first three: "A migration's verify block made of `SELECT`s…", "Deleting a workflow
step…", "Cloudflare answers `Python-urllib` with 403…"). The corrected banner now
lists exactly these, and only these, on every sync run. Fixing them means authoring
footprints, which belongs to the entries' owners.
