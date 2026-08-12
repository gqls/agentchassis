# HANDOFF — concept register — 2026-08-12

> ## ⛔ SUPERSEDED 2026-08-12b — cold-start from `HANDOFF_2026-08-12b_continue_here.md` instead
>
> **Both remaining staleness signals are now closed** (`DOC-078`,
> `scripts/report-register-citation-rot.py`, commit `b9b32ba92`), so this file's staleness
> table sends a session to redo finished work — and its instruction to "try the field key on
> the citation signal first" has been tried: **it does not transfer.** What survives verbatim
> is the enforcement warning below (the clean OPP-006 signal is still not evidence) and the
> lane's landmines, both carried into the new file.

**Cold-start doc for the register lane. This SUPERSEDES
`HANDOFF_2026-08-10b_continue_here.md`.** Two of its items are now closed and its
READ-FIRST banner has been answered twice over — read this instead, not that.

Read after this: `SUMMARY_where_we_are_2026-08-12.md` (the milestone read-out, covering
both the 11th and the 12th), then `FINDINGS_2026-08-10_staleness_survey.md` (**still the live worklist —
its 2026-08-12 UPDATE section closes signal 1 of 3**), then
`FINDINGS_2026-08-11_advisory_delivery.md` (read its corrected banner first), then
`RUNNING_NOTES_concept_register.md` (technical log, newest at the bottom) and
`README_where_we_are.md` (owner prose, append-only).

---

## State in one paragraph

The register is **complete, self-consistent, self-monitoring, gated at authoring time
(advisorily), and the gate is now audible.** Entries and index rows agree exactly; **the
count is deliberately not written here** — it is derived (owner ruling 2026-08-09), so run
the drift check. Staleness — "are the entries still TRUE?" — has moved from surveyed to
**one signal closed and tooled**: version lag is visible via `DOC-077`, with its premise
narrowed by measurement. Two signals remain open.

## What changed on 2026-08-12 (two pieces, both committed)

1. **The advisory-delivery verify-later was BLIND, and it printed a regression.**
   `advisory-delivery-sweep.py --since <day>` read only `toolUseResult.stdout`, so it scored
   every OPP-007 out-of-band delivery as a MISS: **38.2% on a day that was 100%.** Fixed to
   read both channels (`8d74fe75c`). **True reading: 36 of 36 multi-file commits reached
   their session, 23 of them only because of OPP-007.** Controls hold — `oob=0` on every day
   before the hook existed, pre-fix rates unchanged at 55–56%, channel-1 `tail` separation
   intact.
2. **Version lag is built and visible — `DOC-077`** (`scripts/report-register-version-lag.py`,
   `531473d35`). Read-only, ~0.3s, cluster-optional, **not scheduled and not a checker**.
   Corrected `SYS-077`, answered `HITL-020`'s verify-later.

## ⚠ READ THIS BEFORE TOUCHING THE ENFORCEMENT QUESTION

**The case for making OPP-006 blocking is NOT strengthened by anything that has happened —
and the clean signal you will see is not evidence.** The register shows **0 OPP-006 findings
across all 17 register-touching commits** since OPP-007 shipped, and no entry-without-row at
HEAD. That looks like "delivery was the binding constraint, and it worked". **It is a coin
flip:** only **4 entry-adding commits** landed in the window, and at OPP-006's measured 16%
historical leak rate, **P(zero leaks | nothing improved) = 0.50.**

**~14 entry-adding commits are needed for 90% power, ~18 for 95%** — days at this cadence.
Until then: delivery **PROVEN**, behaviour **OPEN**. Use the per-commit sweep, not a HEAD
snapshot (a count at HEAD cannot see a leak repaired the same afternoon).

## The staleness picture — signal 1 of 3 closed

| signal | state |
|---|---|
| **version lag** | **CLOSED 2026-08-12** — tooled as `DOC-077`, premise narrowed, 2 entries corrected |
| **unresolvable `sources:` citations** | **OPEN — 96 of 2,611** (3.7%), mostly the numbered-docs tree deleted 08-04 |
| **moved bug references** | **OPEN — 156.** ⚠ ONE-DIRECTIONAL: owner ruled 08-06 a fixed bug STAYS in `bugs_open/`, so a non-moved bug proves nothing |
| features awaiting a non-roll condition | 5 — `CQ-019` (migration 303), `PLAN-047` (seed 306), `PBP-025` (a `run_checks` array), `TL-038`/`TL-040` (a live fence) |
| roll-conditional class | settled 2026-08-10 via `BLD-019` build provenance |

**The one question to carry into the two open signals** — it is what made version lag
trustworthy: **is there a key that does not require reading prose?** Version lag only worked
once it stopped classifying sentences and keyed on the register's own **field vocabulary**
(`status:` / `status-evidence:` are current-state claims by convention). Classifying by
surrounding words left **77% unclassified**, because `"deployed in chassis v1.0.1029"` (a
permanent fact) and `"both replicas of v1.0.1218 return X"` (an expiring verification) are
indistinguishable by pattern. Try the field key on the citation signal first.

**Still the design law, and DOC-077 obeys it:** report **"this entry's evidence has
expired"**, never "this entry is wrong". Key on things with no prose ambiguity — a version, a
path, a bug id, a date. That restraint is why these checks get read.

**Second cheap move, still not done** (carried from 08-10b, still right): **13 of 29 entries
examined cite NO commit sha**, so provenance can only date them by inference. *An entry whose
status is conditional on a roll must name its commit.* Nine characters when written, a
one-command check for ever. **Candidate for OPP-006, not for a watcher** — put the check where
the error is made.

## How to check the register in 15 seconds

```bash
./scripts/test-concept-register-drift-local.py              # the live check's logic, against HEAD
./scripts/test-concept-register-drift-local.py --self-test  # + historical control and two mutations
./scripts/pattern-check.py                                  # the authoring gate, against staged changes
./scripts/report-register-version-lag.py --worklist         # DOC-077: whose evidence has expired
scripts/advisory-delivery-sweep.py --since 2026-08-13       # is the advisory still reaching people?
```

⚠ The first three read a **ref**, never your worktree; the CronJob reads the **pushed** branch.
So the drift harness **cannot see an entry you have not committed** — verify a new entry/row
pairing by grepping the working tree. Full command set: `RUNBOOK_concept_register.md` §B3, and
**§B11 for the delivery sweep's five gotchas** (two channels, prefix join, UTC vs BST).

## Owed / blocked

- **`rebuild-cascade.md`'s stored count — still owed, FOURTH session running.** Still dirty in
  the shared tree with another session's REB-003 rewrite; mtime unchanged at **2026-08-08
  20:41**, 3 added / 3 deleted. So it is stalled work, not abandoned, and a pathspec commit
  would take it as a same-file passenger. **Re-check `git status` before assuming; if clean,
  retire the line and delete `rebuild-cascade.md` from `KNOWN_STORED_COUNTS` in the local
  harness. Do NOT grow that set to silence findings.** It is the drift check's only HEAD
  finding, so the register otherwise reads clean.
- **A stray `register/model-infrastructure.md.tmp_check`** has been untracked in the tree —
  another session's, left alone.
- **The branch is unpushed.** The watcher reads the **pushed** branch, so its morning row can
  name concepts whose rows are already committed here. Not this lane's call to push.

## Landmines specific to this lane

- **The watcher reads the PUSHED branch**; the harness reads whatever ref you give it. A
  "clean" verdict is never a statement about your working tree.
- **`REGISTER_REF` is hand-pinned** in the manifest (`087_towards_multiple_domains`, correct
  today) and is the *second* such ref in the estate. A stale-but-resolving ref is the worst
  case: every finding unfalsifiable **and every clean run meaningless**.
- **Four commands count the index and all four are correct.** Only
  `grep -cE '^\| [A-Z]{2,4}-[0-9]{3} \|'` is the row count. **Do not write any of them down
  anywhere** — stored counts were retired 2026-08-09 by owner ruling.
- **Two ConfigMaps exist** for the CronJob (kustomize does not prune). Only the mounted one is
  live — diff *that* one against the repo (`RUNBOOK` §B3).
- **A daily report's count is not a RATE.** The watcher only sees what is still broken at
  06:50, so anything fixed the same afternoon never enters its numbers.
- **⚠ NEW 2026-08-12 — an image tag quoted from a live row is stale by construction.** All 187
  live `agent_definitions` rows carry the live tag (uniform), so the release rewrites them.
  **But the identical citation about a repo SEED file is permanent** — `SYS-077` and `HITL-020`
  both cite `v1.0.407` and only one was wrong. Read which artefact holds the tag.
- **⚠ NEW — search live config by the name the ROWS use, not the name the ENTRY uses.** The HITL
  demo agent's type contains no "hitl" (`simple-content-writer-with-approval`) and its group is
  filed under a display name (`Content Approval with HITL`), so both obvious queries return 0
  and read as "never loaded". This session made that mistake before its own control caught it.
- **⚠ NEW — count the paths on your pathspec commit.** 8 named, 7 landed: another session had
  committed `system-architecture.md` in between and took the `SYS-077` correction as a
  passenger. Nothing was lost, but `git log` now attributes it elsewhere.

## Things deliberately NOT done

- ~~**No SUMMARY written.** By the five-headings test, "where we are now" would read
  substantially as 2026-08-10's did — the register's overall state is unchanged.~~
  **SUPERSEDED same day, at the owner's request: `SUMMARY_where_we_are_2026-08-12.md`.** My
  five-headings call was wrong on two counts — 2026-08-11's inaudibility finding never got a
  read-out at all, and both questions the 08-10 summary closed on are now answered, one of them
  by being retired as the wrong question. That is an inflection, not a repetition.
- **No entry's prose was re-verified for truth** beyond `SYS-077` and `HITL-020`.
- **The 13 sha-less entries were not "fixed"** by looking up their commits — the authoring rule
  is the real fix.
- **No `--update-ratchet` run**; the coverage report is quiet.
- **Nothing pushed.**
