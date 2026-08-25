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
