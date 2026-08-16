# HANDOFF 2026-08-16b — what this session finished, and where a fresh chat should start

Written for a stranger with no context. Everything below is committed; nothing is
uncommitted in the tree from this lane.

## 1. Finished, needs nothing further

**`bugs_closed/265` — the legacy `input_schema` dialect is now unrepresentable.** Both halves
live and proven; full evidence in the file's §CLOSED. One-line version: a code comment claimed
the retired JSON-Schema dialect was extinct, four components had been seeded in it since, and
**all four were hand-authored SQL** (not the component-creator, which the bug file had guessed)
— so the fix is a CHECK constraint on `content_components`, the one seam every producer crosses.
Migration `437` applied 10:24Z (converts the last 3 rows behaviour-preservingly, refusal
induced), Go half live on `v1.0.1304` both replicas with a negative control. Council
`aba82416` APPROVED r1. Register `CLC-015`; landmine + `WRONG_CALLS` rows filed.

**`bugs_closed/145` and `bugs_closed/072` — stale duplicates removed.** Both had been closed on
2026-07-31 and *copied* rather than moved, so they sat in `bugs_open/` for a fortnight. Found
by tripping over it: I pod-verified 145's fix before discovering it was already closed. 072's
stale copy held 95 unique lines (a re-verification and a measured correction) — reproduced
verbatim into its closed copy under §RECOVERED before the path was removed. **`bugs_open/` is
106 files now, not 108.**

**New detector:** `scripts/pattern-check.py` `check_bug_file_duplicated` — fires on any commit
touching either bug directory when a basename exists in both. Induced on `a4bf9f1e9`, the very
commit that created the 145 duplicate; control commit silent. Advisory.

## 2. The highest-value next task, already scoped

**Sweep the bugs that are blocked ONLY on a roll against `v1.0.1304`.** A fresh fleet build
landed at 10:41Z today and a class of `bugs_open/` files say, in their own words, "fixed,
committed, not yet live". Each is cheap to settle and some are already closable. Candidate list
(text mentions "not yet live" / "inert until" / "rides the next roll" — **the mention is not the
status; read the header block**):

`040 071 072* 083 085 093 113 117 131 136 145* 151 198 201 204 208 211 220 236` … (`grep -lie
"not yet live" -e "inert until" -e "rides the next" bugs_open/[0-9]*.md`). *Starred ones are
done — the list is from before the duplicates were removed; regenerate it.*

**How to settle one, in order** (this is the recipe that worked today):
1. `ls bugs_open/NNN_* bugs_closed/NNN_*` — two paths for one number means one is stale. **Do
   this before anything else**; it is the check I skipped and it cost a wasted verification.
2. `python3 scripts/who-owns.py NNN` — if OWNED and the lane is active, contribute into the bug
   file rather than closing it under them. "Quiet 14d" plus the file's own stated close
   condition is enough to act.
3. Read the header for the fix commit sha, then:
   `git merge-base --is-ancestor <fix-sha> <the running stamp>`.
   The stamp: `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=400 | grep -m1
   'build provenance'` — **and expect it to be absent**, because it is a startup line on a busy
   service. Fall back to probing the binary for candidate shas:
   `kubectl -n ai-persona-system exec <pod> -- grep -aq "<sha>" /proc/1/exe`.
   Today's chassis stamp is **`5de6cddbe6b281da97dc933d823ebe84da2bbf8a`** (v1.0.1304, pods
   `agent-chassis-5d95ddddfd-48lv6` / `-vtfdx`, started 10:41Z) — reuse it while those pods live.
4. Probe the fix's own literal on **both** replicas, always with a must-be-present control AND a
   must-be-absent control in the same breath. **Never `strings`** — it is not in these images and
   its failure looks exactly like "not found" (145's own file told the reader to use it; that is
   corrected in its closure).
5. Check the fix is per-SERVICE: `grep -rln <symbol> --include=*.go` and see which binaries
   compile it. A chassis roll does not deliver an adapter's fix.
6. If live: append the verification to the bug file, then `git mv` and commit **naming BOTH
   paths**, verifying at HEAD (`git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ |
   grep NNN` → exactly one line). The new pattern-check will tell you if you got it wrong.

## 3. Owed to other lanes, not to be picked up here

- **`report-dossier`'s `body` is `source: llm`** while its seed (`sql_for_agents/207`) says the
  body is never LLM-authored. Migration 437 preserved existing behaviour rather than changing
  it. The honest v2 value is `source: renderer` (134 rows use it). **Gripper-dossier lane's call.**
- **`site-header` carries a third schema shape** — v2 field definitions with no `fields`
  wrapper, which `SchemaContentFields` reads as "no declared fields". Not the retired dialect,
  not refused by 437, harmless today. Noted in `bugs_closed/265` §4.
- **`bugs_open/207` and `bugs_open/217`** read "PROVEN LIVE" in their own headers but sit in
  `bugs_open/`, citing the owner's 08-06 "finished bugs stay" ruling — which the 08-12 ruling
  superseded (fixed-AND-live bugs move again). Their lane (`bugfix_207_sender_convergence`)
  should decide; I did not move another lane's finished bugs on my reading of a ruling.

## 4. Two traps this session paid for, worth inheriting

- **An applied migration is history.** I nearly drifted `437`'s recorded md5 by fixing a stale
  doc pointer in its comment — `schema_migrations` fingerprints the file as applied. Reverted;
  checksum verified equal. Check `SELECT checksum FROM schema_migrations WHERE filename LIKE
  '<NNN>%'` before editing anything under `sql_for_agents/`.
- **A pathspec commit still takes a same-file passenger.** My `WRONG_CALLS.md` commit added 124
  lines where my row was ~30 — two other lanes' entries rode along. Named in
  `NOTES_legacy_dialect_unrepresentable.md` so they can find them. Read `--numstat` after
  committing a contended ledger; "I followed the rule" is not "I checked the result".
