# HANDOFF — `bugs_open/223` lane, cold start for a fresh chat

> **SUPERSEDED IN FULL 2026-08-15 → read `HANDOFF_2026-08-15_continue_here.md`.**
> Everything below was ruled and executed the same day it was written: N=10 ruled,
> the cron built and proven on its first scheduled run, the note-writer review done
> (ack at 11), 223 moved to `bugs_closed/`. This file stays as the record.

Written 2026-08-14. **Supersedes `HANDOFF_2026-08-10_continue_here.md` in full** — both of
that file's live items (§0 items 1 and 2) are DONE and verified. Read this file, then
`NOTES` (newest at the bottom, the 2026-08-14 block) and the bug files it names.
Everything below is verified unless marked otherwise.

---

## 0. WHAT IS ACTUALLY LEFT (read this and you can stop)

**Nothing in this lane is executable by a fresh session without an owner ruling.**
The lane's two bugs (`223`, `254`) are both CLOSED and live. What remains:

1. ~~**OWNER: the RFC_022 budget N.**~~ **RULED 2026-08-14: N = 10** — wired as the
   wrapper's default (`--census` keeps the no-budget mode), both rosters updated
   (migrations 404/405, applied + recorded), RFC_022 STATUS → CLOSED. **The owner also
   corrected the framing, now binding:** sharing is estate design (agents deliberately
   reusable across workflows); a finding means accumulated SURFACE owes one review,
   never that reuse is the problem. Auto-memory: `shared-actions-are-estate-design…`.
   **Follow-through nobody has scheduled yet:** the standing stock at N=10 is exactly
   `analyse_repo_local` (12), `append_doc_note` (11, 8 carriers),
   `diagnose_prepare_fix_commit` (11) — each owes ONE ordinary architecture review of
   its accumulated surface, then its acknowledged level is the baseline.
2. **OWNER (half-ruled): the counter CronJob.** Ruled 2026-08-14: the RFC_006 check's
   Python mirror STAYS (no Go-native rework), and a counter CronJob, if built, may use
   that same shape (mind its landmine: the mirror needs a parity test pinning it to the
   Go source). Whether to BUILD one at all is still open; until then the counter is
   run-on-demand with N=10 default.
3. **WATCH: the as-of note's first live rendering.** 254's fix is live and the persisted
   evidence-line clause is proven end-to-end, but the answer-site as-of note renders only
   on EMPTY answers and none has occurred live yet (build-time it is mutation-proven on
   all three arms; the literal is in the running binary, 1/0 with control). The check:
   `SELECT count(*) FROM llm_call_log WHERE step_name='verify' AND prompt_rendered LIKE
   '%as-of: this answer describes commit%';` — a zero means "no empty answer yet", NOT
   "note missing". When it first fires, note it in `bugs_closed/254`'s closure section.
4. **WATCH (cheap, optional): whether the verifier still manufactures kind-shaped
   explanations now the commit vocabulary sits at the answer site.** `[UNMEASURED]`
   fleet-wide; the both-ways proof is n=1. A natural experiment arrives with every
   landmine filed same-day as its code: compare the verdict's reason against the as-of
   note in the same prompt.

## 1. State in one paragraph

`223` (kinds/extensions): both phases live and proven since v1.0.1284/1286 — the lookup
states what its corpus cannot represent, and Go `var`/`const` are indexed with bodies.
`254` (its commit-axis sibling, found by this lane's own 090 round): the index mirrors
the last PUSHED tip, the staleness caveat lived only in a header the model read and
talked past, and the persisted verdicts were undatable. Fixed by placement: every
in-scope empty answer now carries "as-of: this answer describes commit X (ref, date) …
this 0 is INDEX STALENESS — not absence, not removal, not a rename", and
`codeEvidenceLine` (→ every persisted verdict, both lanes) ends "Answers describe
indexed commit X … not the present tree." Council `42afbd67` approved unanimously; live
on v1.0.1297 (digest + ancestry + binary literal, each with a control); a real
verifier run persisted the dated verdict. RFC_022's counter
(`--optional-key-budget`, register WFA-013) is built, tested, censused against the live
fleet, and both roster clauses (migrations 402/403, applied + recorded) now cite it.

## 2. What is live (do not redo any of this)

| thing | where | proof |
|---|---|---|
| 223 phase 1+2 (census, NOT-ANSWERABLE, var/const with bodies) | see `HANDOFF_2026-08-10_continue_here.md` §2 | unchanged, verified through v1.0.1286 there |
| **254: as-of note on empty answers + dated evidence line** | `diagnose_code_lookup_action.go` (`indexedAsOfNote`, `commitDateClause`, `codeEvidenceLine`), `diagnose_load_runtime_action.go` | commit `0c880908a`; live v1.0.1297; persisted verdict corr `16f0475d` carries the clause |
| freshness struct threaded, not re-queried | `codeIndexFreshness` returns it; `loadCodeIndexScope` takes it as a PARAMETER (forgetting = compile error) | `go build`; tests |
| **RFC_022 counter** | `cmd/config-key-audit/optionalbudget.go` + `scripts/audit-optional-key-budget.sh` | commit (2026-08-14); live census run; register WFA-013 |
| roster clauses name the counter | migrations `402` (fix-proposer 11,829→12,078) + `403` (council-gate 11,866→12,115) | applied via psql + `--record-only`; 403 verify: breakpoint unmoved at 174, **17 seats marked, 1 shared prefix**, cross-roster guard |
| register + paperwork | WFA-013 (entry + index row), RFC_022 STATUS, 016b §9 "a caveat in the HEADER does not protect the ANSWER" + §10 rows 254, `bugs_closed/254` | committed with their code, per the platform-seams ruling |

Councils: `495df717`, `3af67677` (223, both APPROVED); `42afbd67` (254, **approved, all
reviewers**). 090 runs: `520b2f7e` (the staleness round — UNVERIFIABLE with a
premise-refuting last hypothesis; settled first-hand with its own prescribed check).

## 3. The three things a fresh session will most likely get wrong

**(a) Do NOT re-verify 254 with a pod-grep behind `2>/dev/null`.** This lane's first
probe pass returned thirteen clean "absent" rows — including the true revision sha —
because the pods had been deleted mid-probe and every `NotFound` was swallowed. A dead
pod and an unstamped binary are indistinguishable once stderr is gone. Digest-match
first (`docker image inspect` label/digest vs pod imageID), then probe with stderr
visible and BOTH controls. The fleet rolled 1295→1296→1297 inside one day; read the
service you mean, not the tag you remember.

**(b) Do NOT read the as-of-note zero as a failure (or a pass).** It renders only on
empty answers (deliberately not clock-gated — the motivating incident was at 17h). Zero
occurrences in `llm_call_log` means no empty answer has happened since the roll. §0
item 3 has the check and the recording instruction.

**(c) Do NOT run `099_SYNC_gate_roster.py --apply`** — still suspended; it would revert
migration 377 and the 68% caching saving. 402/403 were surgical for exactly this
reason. Its dry-run "drift: all 17" still means the gate is AHEAD of the mirror.

## 4. Standing traps carried forward (unchanged from the 08-10 handoff)

- The indexer reads the REMOTE tip (`git ls-remote origin <branch>`, never
  `rev-parse origin/…`). Re-indexing an unchanged tree is what makes a kind census a
  controlled comparison.
- `landmines-sync.py --apply` consumes the NEEDS_VERIFICATION signal — use
  `./scripts/landmines-verify-dispatch.sh` when you want changed entries verified.
- No orchestration dispatch within ~300s of a chassis pod (re)start.
- `verify_unverifiable` has still never executed in production (`no_code_evidence`
  keeps not happening); do not credit the gate with the evidence layer's work.
- `code_symbols` has `line_start`/`line_end`, not `start_line`; no `indexed_at`.
- Never `ORDER BY created_at DESC LIMIT 1` on a shared table — filter by correlation.
- `go run` folds a tool's exit 2 into 1 — wrappers discriminate refusal by EMPTY
  STDOUT (`audit-optional-key-budget.sh` shows the pattern).

## 5. Commands

```bash
# what is the service running (per SERVICE; stamp scrolls — label+digest is the fallback)
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}' docker.io/aqls/agent-chassis:<tag>
git merge-base --is-ancestor <commit> <revision> && echo LIVE

# the counter (census; add N for findings + exit 1)
./scripts/audit-optional-key-budget.sh [--json] [N]

# fire one landmine entry at the verifier
./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>' <branch>
```
```sql
SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 1;   -- corpus census
SELECT DISTINCT commit_sha, ref FROM code_symbols;               -- what the index has seen
```

## 6. Where the paperwork is

This directory: PLAN, NOTES (append-only; 2026-08-14 block is the latest), RUNBOOK,
README_where_we_are (owner's log — the 08-14 entry states his two open decisions),
SUMMARYs 08-10/08-11, three council submission JSONs (223 ×2, 254 ×1), the superseded
08-10 handoff. Fleet-wide: `bugs_closed/223`… wait — **223 is in `bugs_open/` with a
closed banner (owner 08-06 era) and 254 is in `bugs_closed/` (owner 08-12 bar)**; the
08-12 restoration means 223 could now also move, but that is its own small decision —
do not move it silently, note it to the owner instead. `016b` §9 (two entries from this
lane) + §10 rows, `WRONG_CALLS.md` (3 rows), `LANDMINES.md` (staleness entry, corrected
08-11), `architecture_review/RFC_022_*` (STATUS = counter built, N awaited).
