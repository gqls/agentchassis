# HANDOFF — concept register — 2026-08-12b

**Cold-start doc for the register lane. This SUPERSEDES
`HANDOFF_2026-08-12_continue_here.md`** (which superseded `2026-08-10b`). Its staleness
table is now out of date in two rows — both remaining signals are closed — but **its
enforcement warning still stands verbatim and is repeated below, because nothing that
happened since strengthens that case either.**

Read after this: `FINDINGS_2026-08-10_staleness_survey.md` — **read its 2026-08-12b UPDATE
at the bottom first**, then the 08-12 one above it — then
`SUMMARY_where_we_are_2026-08-12.md` (the milestone read-out; still current, see "no new
summary" below), then `RUNNING_NOTES_concept_register.md` (technical log, newest at the
bottom) and `README_where_we_are.md` (owner prose, append-only).

---

## State in one paragraph

The register is **complete, self-consistent, self-monitoring, gated at authoring time
(advisorily), and the gate is audible.** Entries and index rows agree exactly; **the count
is deliberately not written here** — it is derived (owner ruling 2026-08-09), so run the
drift check. Staleness — "are the entries still TRUE?" — is now **surveyed, tooled and
closed on all three signals**: `DOC-077` for version lag, `DOC-078` for citations and moved
bug references. What remains open on staleness is not a signal but a **question about
uptake**: two reports now exist that nobody is yet obliged to read.

## What changed on 2026-08-12b (one piece, committed `b9b32ba92`)

**`DOC-078` — `scripts/report-register-citation-rot.py`.** Read-only, ~1.5s, no cluster,
no DB, **not scheduled and not a checker**. Closes staleness signals 2 (unresolvable
`sources:` citations) and 3 (moved bug references) with one resolver that speaks the
register's own citation forms. `--self-test` (10 cases) · `--worklist` · `--list <VERDICT>`.

**The result inverts the survey's picture.** 7,793 path citations across 1,767 entries:
**75% resolve as written**, 801 name their own repair (the file is at HEAD elsewhere, or in
git, or the bug moved), 769 are under-specified (a bare filename matching several files —
vague, never wrong), 345 are declared unjudgeable, and **4 name a file that has never
existed.** Three of those four sit in `verify-later:`. The 08-10 figure of "96 of 2,611"
was right for the question it asked (does this path exist at HEAD?); asking instead what
git can still find moves nearly all of it. **The citations are not rotting — they are
abbreviated, and one directory move on 2026-08-04 broke a lot of abbreviations at once.**

## The carried question is ANSWERED, and the answer is no

The handoff you are replacing said: *"is there a key that does not require reading prose?
Try the field key on the citation signal first."* **Tried; it does not transfer.**
Unresolved rates run 10–37% across every field with no break, because a citation's field
predicts nothing about whether its target was renamed. Version lag was the only signal with
that shape.

**What does work is a different structural key — what git can say about the target**
(at HEAD / moved / deleted / never existed). Total, mechanical, and the middle verdicts
name their own repair. **The field key comes back as SEVERITY rather than as a filter:** a
dead path in `sources:` is a grounding claim nobody can open; the same path in
`verify-later:` is a to-do named wrongly. The report prints that ordering.

**The design law is unchanged and both reports obey it:** say *"this citation does not
resolve as written"*, never *"this entry is wrong"*.

## ⚠ READ THIS BEFORE TOUCHING THE ENFORCEMENT QUESTION (carried forward unchanged)

**The case for making OPP-006 blocking is NOT strengthened by anything that has happened —
and the clean signal you will see is not evidence.** 0 OPP-006 findings across the
register-touching commits since OPP-007 shipped, and no entry-without-row at HEAD. That
looks like "delivery was the binding constraint, and it worked". **It is close to a coin
flip:** only a handful of entry-adding commits have landed in the window, and at OPP-006's
measured 16% historical leak rate, four clean ones happen half the time by luck alone.
**~14 entry-adding commits are needed for 90% power, ~18 for 95%.** Until then: delivery
**PROVEN**, behaviour **OPEN**. Use the per-commit sweep, not a HEAD snapshot.

## What is actually open now

| item | state |
|---|---|
| **the 801 repairable citations** | **not repaired, deliberately.** An automated 801-line rewrite across 111 files is the change no reviewer can check, and each citation was correct when written. `DOC-078`'s `verify-later` asks whether anyone repairs any by hand; if the answer in a month is "nobody", the fix is a `sources:` convention at authoring time, not a louder report |
| **the 4 dead citations** | `ADP-018` (`sources:`, the sharp one — right bug number, wrong dir/date/slug), `VET-006`, `SYS-004`, `HITL-017` (all `verify-later:`). Listed with their real targets in the FINDINGS 08-12b update. **Not corrected here** — each belongs to a lane that knows what it meant |
| **the sha-citation authoring rule** | **still not done, and still the cheapest thing on this list.** 13 of 29 entries examined cite no commit sha, so provenance can only date them by inference. *An entry whose status is conditional on a roll must name its commit.* Nine characters when written, a one-command check for ever. **Candidate for OPP-006, not for a watcher** — put the check where the error is made |
| **uptake of the two reports** | the real open question. `DOC-077` and `DOC-078` are both unscheduled by design. If nothing is ever acted on, the honest conclusion is that authoring-time gates work here and reports do not — which is a finding, not a failure |
| features awaiting a non-roll condition | 5 — `CQ-019` (migration 303), `PLAN-047` (seed 306), `PBP-025` (a `run_checks` array), `TL-038`/`TL-040` (a live fence) |

## How to check the register in 15 seconds

```bash
./scripts/test-concept-register-drift-local.py              # the live check's logic, against HEAD
./scripts/test-concept-register-drift-local.py --self-test  # + historical control and two mutations
./scripts/pattern-check.py                                  # the authoring gate, against staged changes
./scripts/report-register-version-lag.py --worklist         # DOC-077: whose evidence has EXPIRED
./scripts/report-register-citation-rot.py                   # DOC-078: whose evidence cannot be OPENED
./scripts/report-register-citation-rot.py --self-test       # 10 cases, each naming the wrong answer it guards
scripts/advisory-delivery-sweep.py --since 2026-08-13       # is the advisory still reaching people?
```

⚠ The drift trio read a **ref**, never your worktree; the CronJob reads the **pushed**
branch. **`DOC-078` is the exception — it reads the working tree**, so it sees an entry you
have not committed and the drift harness does not. Full command set:
`RUNBOOK_concept_register.md` §B3, and §B11 for the delivery sweep's five gotchas.

## Owed / blocked

- **`rebuild-cascade.md`'s stored count — still owed, FIFTH session running, and the reason
  CHANGED on 2026-08-12b.** It is no longer dirty in the tree: a session ran `git stash` at
  18:38:52, sweeping 38 files' uncommitted work — **including this one's REB-003 rewrite** —
  into `stash@{0}`. So `git status` now reads clean and the file looks free to edit. **It is
  not.** The owning session can restore that stash at any time, and editing the file now sets
  up a conflict they will have to resolve. ⚠ **Do NOT `git stash pop`** to inspect it — that
  drops all 38 files into the shared tree. `git stash show --name-only stash@{0}` is read-only
  and answers the question. Stalled, not abandoned; a pathspec commit would take it as a
  same-file passenger. **Re-check `git status` before assuming; if clean, retire this line
  and delete `rebuild-cascade.md` from `KNOWN_STORED_COUNTS` in the local harness. Do NOT
  grow that set to silence findings.** It is the drift check's only HEAD finding.
- **A stray `register/model-infrastructure.md.tmp_check`** is still untracked — another
  session's, left alone.
- **The branch is unpushed.** The watcher reads the **pushed** branch, so its morning row
  can name concepts whose rows are already committed here. Not this lane's call to push.

## Landmines specific to this lane

- **The watcher reads the PUSHED branch**; the harness reads whatever ref you give it. A
  "clean" verdict is never a statement about your working tree.
- **`REGISTER_REF` is hand-pinned** in the manifest (`087_towards_multiple_domains`,
  correct today) and is the *second* such ref in the estate. A stale-but-resolving ref is
  the worst case: every finding unfalsifiable **and every clean run meaningless**.
- **Four commands count the index and all four are correct.** Only
  `grep -cE '^\| [A-Z]{2,4}-[0-9]{3} \|'` is the row count. **Do not write any of them down.**
- **Two ConfigMaps exist** for the CronJob (kustomize does not prune). Only the mounted one
  is live — diff *that* one against the repo (`RUNBOOK` §B3).
- **A daily report's count is not a RATE.** The watcher only sees what is still broken at
  06:50; anything fixed the same afternoon never enters its numbers.
- **An image tag quoted from a live row is stale by construction** (all 187 live
  `agent_definitions` rows carry the live tag), **but the identical citation about a repo
  SEED file is permanent.** `SYS-077` and `HITL-020` both cite `v1.0.407`; only one was
  wrong. Read which artefact holds the tag.
- **Search live config by the name the ROWS use, not the name the ENTRY uses.** The HITL
  demo agent's type contains no "hitl" (`simple-content-writer-with-approval`) and its
  group is filed under a display name, so both obvious queries return 0 and read as "never
  loaded".
- **⚠ NEW 2026-08-12b — `000_concept_index.md` should be EXPECTED to ride out under another
  session's commit, not merely feared.** It happened twice in two sessions
  (`4a6e39c28` took a `system-architecture.md` correction; `11abe7a41` took the `DOC-078`
  row, leaving a row at HEAD promising a concept no file defined). **The tell is arithmetic:
  count the paths you named against the files in the commit.** Then commit the other half
  promptly — the window where a row exists without its entry is the exact state OPP-006 is
  built to prevent.
- **⚠ NEW 2026-08-12b — `git rev-list --objects --all` CANNOT enumerate paths.** It dedups
  by object, so content-identical files share one blob and only one path is printed: **791
  of 9,301 HEAD paths absent, all duplicates.** Every dropped path reads as "this file has
  never existed", which in a citation check is a *finding*, not a gap. Use
  `git log --all --no-renames --pretty=format: --name-only`, and assert `HEAD ⊆ ever`
  before reporting. Fleet-wide in `LANDMINES.md`.
- **⚠ NEW 2026-08-12b — normalising a citation before resolving it MANUFACTURES findings.**
  The `(N)` suffix is an extraction-unit id in some citations and **part of the real
  filename** in others. Stripping it unconditionally produced 27 of 34 "never existed"
  findings, sorted to the top by frequency, and one of them was stated aloud before it was
  checked. **Resolve what was written before resolving a guess at what was meant**, take the
  BEST verdict across variants rather than the last, and **never print an absence without
  the near-miss git does have.** `WRONG_CALLS.md`, 2026-08-12.
- **⚠ NEW 2026-08-12b — `Council-Submitted: n/a` is refused by the commit-msg gate**, and it
  is right to: the trailer is a join key for the 098 report and a non-UUID resolves to
  nothing. Out-of-scope change (`scripts/`, docs)? **Omit the trailer** and say so in prose.

## Things deliberately NOT done

- **No new SUMMARY.** By the five-headings test this does not clear the bar:
  `SUMMARY_where_we_are_2026-08-12.md` was written hours ago and its "where we're going"
  named these two signals as the next work. Closing them changes "where we are" by one line
  and would repeat the rest almost verbatim. The material is in the FINDINGS update, NOTES
  and `README_where_we_are`. **The next inflection is uptake** — whether anyone acts on
  either report — and that is days away at the earliest.
- **None of the 801 repairable citations were rewritten**, and the four dead ones were left
  for their own lanes.
- **The 13 sha-less entries were not "fixed"** by looking up their commits — the authoring
  rule is the real fix.
- **No `--update-ratchet` run**; the coverage report is quiet.
- **Nothing pushed.**
