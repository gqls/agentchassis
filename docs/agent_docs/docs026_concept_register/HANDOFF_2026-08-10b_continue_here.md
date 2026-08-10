# HANDOFF — concept register — 2026-08-10b (evening)

**Cold-start doc for the register lane. This SUPERSEDES
`HANDOFF_2026-08-10_continue_here.md`** (that one's items 1 and 3 are both done;
its rate figure is corrected below). Written because the chat grew long, not
because the work stalled.

Read after this: `SUMMARY_where_we_are_2026-08-10.md` (milestone),
`FINDINGS_2026-08-10_staleness_survey.md` (**the live worklist — read this before
doing anything on staleness**), then `RUNNING_NOTES_concept_register.md` (technical
log, newest at the bottom) and `README_where_we_are.md` (owner prose, append-only).

---

## ⚠ READ FIRST — the gate's first live test, and it did not pass

**OPP-006 fired and was ignored, within three hours of shipping.** The gate went
into the tree at **19:08**; commit `5c7b115c5` at **20:31** added `DES-082` and
`DES-083` with no index rows, and `./scripts/pattern-check.py --commit 5c7b115c5`
confirms it names both. Rows have been backfilled (`95429cbe8`).

This is exactly the alternative outcome OPP-006's own **verify-later** names: *if
the watcher's missing-row count does not fall to zero, the gate is being ignored
rather than working, and that is a different problem with a different fix.*

**Do not over-read it, and do not act on it yet.** n = 1 commit, one session,
three hours in — a data point, not a rate. This lane was already caught once
today quoting a rate off an instrument too small to carry it. **Watch the daily
row for a week** before anyone argues for teeth. What it does establish: the
advisory posture has a ceiling — the check cannot make anyone read it, which is
the same ceiling the daily report has, moved earlier in time. The counter-argument
is in `pattern-check.py`'s own docstring and is strong: a check that blocks on a
bad day gets disabled permanently, and a false positive that blocks is a
fleet-wide outage on a shared tree.

**The honest question to answer first is which of these it was** — the session
never saw the output, or saw it and judged the row could wait. Those have opposite
fixes (delivery vs. enforcement), and nothing recorded so far distinguishes them.


## State in one paragraph

The register is **complete, self-consistent, self-monitoring, and now gated at
authoring time (advisorily — see the banner above)**. Entries and index rows agree
exactly; **the count is deliberately not written here** — it is derived, per the
owner ruling of 2026-08-09, and a figure in a handoff is exactly the artefact that
retirement removed. Run the drift check. A daily CronJob
(`concept-register-drift-check`, DOC-074) reports drift; a pre-commit check
(`check_register_entry_without_row`, **OPP-006**, added today) stops the commonest
drift being written at all. Staleness — "are the entries still TRUE?" — has now
been **surveyed and partly settled**: the roll-conditional class is closed, three
other signals are measured and open.

## What was done today (2026-08-10, after the previous handoff)

1. **Built the authoring gate — handoff item 1.** `check_register_entry_without_row`
   in `scripts/pattern-check.py`, run by `.githooks/pre-commit`, registered as
   OPP-006. Two arms: an entry with no index row; an id already claimed. Measured
   over 398 register-touching commits — 84% of entry-adding commits already comply
   and stay silent, 16% leak, 0 false positives, **median leaked entry waited 93
   hours** for its row. Commit `7db343ee7`.
2. **Backfilled the two owed rows** (`BLD-018`, `DIAG-042`) — `a332522df`.
3. **Filed a fleet-wide landmine** — a git helper that captures stdout only turns a
   fatal into an empty result — `728d7d891`, synced to `doc_notes`.
4. **Surveyed staleness — handoff item 3.** `FINDINGS_2026-08-10_staleness_survey.md`,
   commit `8e3c90d1f`.
5. **Settled the survey's worklist using BLD-019's build provenance**, which went
   live on the evening's roll — `ebaac39c0`. 19 entries annotated.

## The staleness picture — one section, because it is the live work

**Settled today:** the "inert until the next chassis roll" class. Chassis
`v1.0.1283` carries BLD-019's stamp; both replicas return
`d3c09cc746e563b6339831cfb69576eb52135c43`, no `-tree` suffix. So:

```bash
git merge-base --is-ancestor <the entry's own commit> d3c09cc746e563b6339831cfb69576eb52135c43
```

⚠ **Control it before you trust it** — an ancestry test that always says yes says
nothing. Positive: `3a59b5012` (FIX-055) → IN. Negative: `3ac87646a` (off-branch
merge) → NOT IN.

**Still open, and untouched by any roll:**

| signal | size | note |
|---|---|---|
| **version lag** | **129 entries cite a chassis version; 80 are 50+ behind** (`SYS-077`/`HITL-020` cite v1.0.407 — 873 rebuilds ago) | the cleanest mechanical signal in the register; needs no prose parsing |
| **unresolvable `sources:` citations** | **96 of 2,611** (3.7%) | mostly the numbered-docs tree deleted 08-04 |
| **moved bug references** | **156** entries cite `bugs_open/NNN` now in `bugs_closed/` | ⚠ ONE-DIRECTIONAL — owner ruled 08-06 that a fixed bug STAYS in `bugs_open`, so a non-moved bug proves nothing |
| features awaiting a non-roll condition | 5 | `CQ-019` (migration 303), `PLAN-047` (seed 306), `PBP-025` (a `run_checks` array), `TL-038`/`TL-040` (a live fence) |

**Cheapest next move, and it is not a checker:** make **version lag** visible. 129
entries already carry the number; the fleet's version is one `kubectl` call.

**Second, and cheaper still — an authoring rule with real leverage:** **13 of the 29
entries examined cite NO commit sha**, so provenance can only date them by
inference. *An entry whose status is conditional on a roll must name its commit.*
Nine characters when written; a one-command check for ever. **Candidate for OPP-006,
not for a watcher** — same argument as the missing row: put the check where the
error is made.

### The design conclusion — do not skip this before building anything

**A staleness checker must NOT parse the `status:` field.** Measured: a regex said
38 entries still claim not-live; reading all 38 said ~20. Not a weak pattern — the
field does four things no pattern can classify:

- `WFA-006` — **"runtime-inert BY DESIGN"**: permanent, reads exactly like expiring.
- `VONC-011` — quotes its own stale wording *inside* the correction that fixed it
  (the frozen-log trap that already forced `check.py`'s searches to be head-bounded).
- `CLC-013`, `STY-056`, `WFA-009`, `CGV-031` — **half live, half not**: one entry,
  two statuses, two clocks.
- `PBP-037` — a chain of three preconditions **in order**, not a state.

Key on things with no prose ambiguity — **a version, a path, a bug id, a date** —
and report **"this entry's evidence has expired"**, never "this entry is wrong".
That restraint is why the drift check is trusted enough to be read.

## How to check the register in 15 seconds

```bash
./scripts/test-concept-register-drift-local.py              # the live check's own logic, against HEAD
./scripts/test-concept-register-drift-local.py --self-test  # + historical control and two mutations
./scripts/pattern-check.py                                  # the authoring gate, against staged changes
```
⚠ Both read a **ref**, never your worktree; the CronJob reads the **pushed** branch.
Full command set with gotchas: `RUNBOOK_concept_register.md` **§B3**.

## Owed / blocked

- **`rebuild-cascade.md`'s stored count — still owed, third session running.** The
  file has been dirty in the shared tree with another session's REB-003 rewrite;
  last written **2026-08-08 20:41**, so that is active work, not an abandonment,
  and a pathspec commit would take it as a same-file passenger. **Re-check
  `git status` before assuming; if clean, retire the line and delete
  `rebuild-cascade.md` from `KNOWN_STORED_COUNTS` in the local harness. Do NOT
  grow that set to silence findings.**
- **The branch is unpushed** (65+ commits across all sessions as of this evening).
  The watcher reads the **pushed** branch, so its morning row can name concepts
  whose rows are already committed here. Not this lane's call to push.

## Landmines specific to this lane

- **The watcher reads the PUSHED branch**; the harness reads whatever ref you give
  it. A "clean" verdict is never a statement about your working tree.
- **`REGISTER_REF` is hand-pinned** in the manifest (currently
  `087_towards_multiple_domains`, correct today) and is the *second* such ref in the
  estate. A stale-but-resolving ref is the worst case: every finding becomes
  unfalsifiable **and every clean run becomes meaningless**.
- **`git fetch` with no refspec** sets `FETCH_HEAD` to whatever branch it wrote
  last. Use `git ls-remote origin refs/heads/<branch>`.
- **Four commands count the index and all four are correct.** Only
  `grep -cE '^\| [A-Z]{2,4}-[0-9]{3} \|'` is the row count. **Do not write any of
  them down anywhere** — stored counts were retired 2026-08-09 by owner ruling.
- **Two ConfigMaps exist** for the CronJob (kustomize does not prune). Only the
  mounted one is live — diff *that* one against the repo (`RUNBOOK` §B3). The
  "redeploy owed" note from 08-09 was already settled when checked this way.
- **`git grep <pattern> --cached` dies** and a stdout-only helper turns that into
  "nothing found" — this shipped inert once and passed a 398-commit sweep. Options
  BEFORE the pattern; a revision AFTER it. `LANDMINES.md`, 2026-08-10.
- **A daily report's count is not a RATE.** The watcher can only see what is still
  broken at 06:50, so anything fixed the same afternoon never enters its numbers.
  This corrected the leak rate from "~1 per 1.5 days" to **~1.2/day**.

## Things deliberately NOT done

- **No entry's prose was re-verified for truth.** The 19 annotations claim only
  that the Go code is in the running binary — never that the feature is exercised.
- **The 13 sha-less entries were not "fixed"** by looking up their commits. Their
  annotation says the inclusion is inferred; the authoring rule is the real fix.
- **No `--update-ratchet` run**; the coverage report is quiet.
- **Nothing pushed.**
