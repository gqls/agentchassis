# HANDOFF — concept register — 2026-08-17

**Cold-start doc for the register lane. This SUPERSEDES `HANDOFF_2026-08-16_continue_here.md`**
(which superseded `2026-08-12b`). Everything below was re-measured at HEAD `7d832ebc8` on
2026-08-17 12:32–12:40 BST, and re-confirmed unchanged at `b4db98f0b` twenty minutes later —
the tree moved under the measurement while I wrote this, which is the normal condition here. **Every figure here carries its command — re-run before repeating
any of them outward; this tree moves ~250 commits a day.**

Read after this, only as needed: `RUNNING_NOTES_concept_register.md` (technical log, newest at
the bottom — §§ 2026-08-12c, 08-14/08-16 and its two postscripts cover the last three sittings),
`README_where_we_are.md` (owner prose, append-only), and `HANDOFF_2026-08-12b_continue_here.md`
**only** for the long-form account of `DOC-077`/`DOC-078`'s design. You should not need the
earlier handoffs at all; their live content is consolidated here.

---

## Start here — the lane is in a QUIET state, and this is the ranked next work

Nothing is broken and nothing is mid-flight. There is no owed read-back, no dispatch in the air.
In priority order:

1. **The sha-citation authoring rule — still the cheapest thing on the list, three handoffs
   running.** 13 of 29 entries examined cite no commit sha, so provenance can only date them by
   inference. *An entry whose status is conditional on a roll must name its commit.* Nine
   characters when written, a one-command check for ever. **Belongs in `OPP-006`
   (`scripts/pattern-check.py`), not in a watcher** — put the check where the error is made.
   This is a small, self-contained, high-leverage piece and it is the one I would pick up.
2. **Test whether the two reports are ACTIONABLE, by acting on one.** This is the lane's real
   open question (below) and it is answered by doing, not measuring: take one entry from
   `./scripts/report-register-version-lag.py --worklist`, repair its expired evidence by hand,
   and record how long it took and whether the report pointed you at the right thing. `DOC-078`'s
   own `verify-later` asks exactly this. A month of "nobody acted on it" is a finding — but only
   if somebody tried once.
3. **Leave `rebuild-cascade.md` alone** unless `git log -1` on it has moved (see Owed).
4. **Do NOT touch the enforcement question** without reading its section below first.

## State in one paragraph

The register is complete, self-consistent, self-monitoring, gated at authoring time (advisorily),
audible, and all three staleness signals are closed and tooled (`DOC-077` version lag, `DOC-078`
citations + moved bug references). The landmine delivery path — the mechanism that carries this
lane's warnings and every other lane's — was found to be keyed so that nothing could match it,
and is now repaired and holding. The shared tree has a mechanical ban on the one command that
erased two days of every lane's uncommitted work. **The count of entries is deliberately not
written here** (owner ruling 2026-08-09): it is derived, so run the drift check.

## Verified at HEAD `7d832ebc8`, 2026-08-17 — with the command for each

| what | reading | command |
|---|---|---|
| register self-consistency | **1880 entries, 1880 index rows — agree exactly** | `./scripts/test-concept-register-drift-local.py` |
| register drift findings | **1** — `rebuild-cascade.md`'s stored count (see Owed) | same |
| landmine rows ↔ file | **546 entries both sides, 0 key mismatches, 0 strays, 0 `·` left in any `subject_key`**; 2,758 rows. ⚠ A lane that appended minutes ago and has not run the dispatcher shows as `NOT IN DB` — that is a normal transient (one was live as I wrote this), not a regression; it is theirs to sync | `./scripts/landmines-keys-check.py` |
| the splitter fix holding for NEW arrivals | **64 entries added since 08-14, all keyed correctly** | `./scripts/landmines-keys-check.py` |
| the stash ban | **no new stash in 650 commits since the ban** (1,043 since the stash itself); `stash@{0}` is still the 2026-08-12 18:38:51 one | `git log -g --pretty='%gd %ad' --date=iso refs/stash \| head -1` |
| verifier verdicts (3 entries) | **UNVERIFIABLE ×3, stable, no objections** — the `verify_unverifiable` branch, correct for wholly non-Go footprints | query in "What changed", below |
| live fleet | **`v1.0.1305`** on 18 deployments; makefile `IMAGE_TAG` matches; 3 distinct tags live | `kubectl -n ai-persona-system get deploy -o jsonpath=…` |

## The register's own open items — re-measured today, not copied forward

**Citation health (`DOC-078`) — stable against a growing corpus, which is the useful reading.**
8,279 path citations across 1,806 entries: **6,248 resolve as written (75%)**, **881 name their
own repair** (392 `BUG-MOVED`, 288 `MOVED-AT-HEAD`, 196 `DELETED`, +5), **778 `MOVED-AMBIGUOUS`**
(the file exists; the citation does not say which one), 43 `UNJUDGED-DIRSHAPE`, and **4
`NEVER-REPO-PATH`** — no file, ever, under that name. Against 08-12's 7,793 citations / 1,767
entries: the corpus grew ~490 citations, **the 75% ratio did not move, and "never existed" is
still exactly 4.** So the citations are not rotting faster than the register grows — they are
abbreviated. The 881 are still **deliberately not repaired**: an automated rewrite across 111
files is the change no reviewer can check, and each citation was correct when written.
The 4 dead ones are `ADP-018` (`sources:` — the sharp one), `VET-006`, `SYS-004`, `HITL-017`
(all `verify-later:`); each belongs to a lane that knows what it meant.

**Version lag (`DOC-077`) — there IS a worklist now, and its controls pass.** 345 version
citations across 182 entries: `status` median lag 36 (49 entries ≥50 behind), `status-evidence`
median lag 111 (**77 entries ≥50 behind**), `verify-later` 51, `what` 115. Controls: newest
citation `v1.0.1305`, lag 0 — so the live version is being resolved correctly; field-keying
excluded 101 of 345 citations (29%), so **the key is doing work rather than passing everything
through.** This is the report earning its place, and item 2 at the top of this doc is the
experiment it is waiting for.

**Uptake is the real open question**, unchanged: both reports are unscheduled by design. If
nothing is ever acted on, the honest conclusion is that authoring-time gates work here and
reports do not — a finding, not a failure. Item 2 is how to stop guessing.

**Features awaiting a non-roll condition: 5** — `CQ-019` (migration 303), `PLAN-047` (seed 306),
`PBP-025` (a `run_checks` array), `TL-038`/`TL-040` (a live fence).

## ⚠ THE ENFORCEMENT QUESTION — read before touching `OPP-006`'s blocking behaviour

**The case for making `OPP-006` blocking is not strengthened by a clean signal, and the clean
signal you will see is not evidence.** ~14 entry-adding commits are needed for 90% power at the
measured 16% historical leak rate, ~18 for 95%. Use the per-commit sweep, never a HEAD snapshot —
a count at HEAD cannot see a leak repaired the same afternoon. Delivery **PROVEN**, behaviour
**OPEN**.

**New evidence since, and it points BOTH ways.** On the morning of 08-16 the "index row rides out
under another session's commit" mechanism fired **twice within minutes** (`PUB-005`, then a
second row) — and **both self-healed inside ten minutes** as the owning lanes committed their
entry halves. So leaks are real and frequent, *and* the repair arrives in minutes without a gate.
Do not read either half alone as settling it.

## What changed 2026-08-14 → 08-17 (all committed; nothing outstanding)

| commit | what |
|---|---|
| `371317eb6` | **`git stash` FORBIDDEN (owner ruling 08-14) and mechanically blocked** — `scripts/block-git-stash.py`, a `PreToolUse` hook in the versioned `.claude/settings.json`; rule in CLAUDE.md § Git. Git has no pre-stash hook, so the harness is the only enforceable layer. Mutating forms denied in any compound shape; `git stash list`/`show` deliberately allowed (they are the documented recovery). 14/14 self-test; proven live by a denied attempt. Same commit carried the owner's 08-12 CLAUDE.md note, which had sat uncommitted — and stash-sweepable — for two days. |
| `f92e0b3ca` | **`split_footprints` fixed; 185 of 482 landmine entries re-keyed.** `·` splits unconditionally; `,`/`;` respect parentheses; the trailing-qualifier strip needs a space before `(` so a SQL signature survives. Also replaced `landmines-sync.py`'s **count**-based rewrite detection with **identity** — 6 entries changed keys at an unchanged count and would have stayed stale for ever, every sync reporting clean. |
| `0b6831dcf`, `d2f7651fe` | lane docs: NOTES for both sittings, README prose, the fresh-at-the-time handoff, the `PUB-005`/transient reading rule. |
| verifier | fired 08-16 ~10:03Z for the three entries this lane owed it (two had been armed by `--apply` on 08-12 and never dispatched). All three spawned within 20 s; **UNVERIFIABLE ×3**, each stating the entry is internally consistent. Read back: `SELECT left(subject_key,46), created_at::timestamp(0), substring(body from 'landmine-verifier\): ([A-Z_]+)') FROM doc_notes WHERE categories ? 'landmine-verification' AND subject_key LIKE 'LANDMINES.md#%' ORDER BY created_at DESC LIMIT 6;` |

The 08-12 incident itself (a bare `git stash` swept 38 files across ~10 lanes and reverted 18
production overlays ~100 releases while leaving the tree looking clean) is written up in
`LANDMINES.md` — grep `git stash` — and in `WRONG_CALLS.md` 2026-08-12 (this lane's "clean read
as RESOLVED" misread). All 38 files were restored and verified per file; `stash@{0}` was left in
place and still holds the un-restored deletion halves.

## Landmines specific to this lane

- **`landmines-sync.py --check` says "in sync" WITHOUT checking key identity.** Its drift test is
  `new or gone` — presence of the source key only. An entry whose FOOTPRINTS changed passes
  `--check` while its `doc_notes` `subject_key`s are stale, which is the exact failure the 08-14
  fix repaired 185 instances of. **The real assertion is `./scripts/landmines-keys-check.py`**
  (built 2026-08-17 for exactly this; read-only, exits 1 on mismatch, and `--self-test`
  mutates a copy of the corpus to prove it can fail). Run that, not `--check`, whenever you
  care whether the keys are right — and note its self-test caught a no-op mutation of its own
  on first run: `parse()` starts at the `# Entries` marker, so a mutation above it registers
  nothing and would have made the whole self-test pass vacuously.
- **A `NEEDS_VERIFICATION` line from `landmines-sync.py --apply` is an UNSENT dispatch.** Two of
  this lane's entries sat armed-and-unsent from 08-12 to 08-16. Run
  `./scripts/landmines-verify-dispatch.sh` (sync **and** dispatch), or
  `./scripts/trigger-landmine-verifier.sh '<slug>'` per entry afterwards. CLAUDE.md was corrected
  on this 08-15.
- **A landmine footprint must be tested for MATCHING, not for syncing** — parse it and assert a
  LIST of short grep-able strings. The `·` convention is legal now, but **a glob still never
  substring-matches a real path**: lead with a plain directory prefix.
  `python3 scripts/landmines_lib.py` is an 8-case self-test; run it before touching
  `split_footprints` (my own first fix failed it).
- **A `verify_unverifiable` / UNVERIFIABLE verdict is the branch WORKING, not a refutation** — the
  verifier's `code_symbols` index holds Go only, so any wholly non-Go footprint lands there. And
  **`NEEDS_HUMAN_REVIEW` may be code-index staleness, not a refutation** (`9f619d938`).
- **A row-without-entry drift finding YOUNGER THAN ~10 MINUTES is a lane mid-commit, not a leak.**
  Measured twice on 08-16; both self-healed. **Re-run the drift check before filing anything.**
- **The drift trio read a REF, never your worktree; `DOC-078` is the exception and reads the
  working tree.** So the harness cannot see an entry you have not committed.
- **The CronJob reads the PUSHED branch, and `REGISTER_REF` is hand-pinned** in
  `deployments/kustomize/services/concept-register-drift-check/base/cronjob.yaml` (the *second*
  such ref in the estate; no default). **CORRECTION 2026-08-17 — prior handoffs said "the branch
  is unpushed"; that is wrong today.** `origin/087_towards_multiple_domains` exists at
  `896c5aeeb` and is **66 commits behind HEAD**, so the watcher's morning row is real but lags
  HEAD by hours. `git status -sb` shows no upstream because none is *configured locally* — that
  is a tracking-config fact, not a statement about the remote. Check it directly:
  `git ls-remote --heads origin 087_towards_multiple_domains` then
  `git rev-list --count <sha>..HEAD`.
- **Two ConfigMaps exist** for the CronJob (kustomize does not prune) — only the mounted one is
  live; diff *that* one (`RUNBOOK` §B3).
- **Four commands count the index and all four are correct.** Only
  `grep -cE '^\| [A-Z]{2,4}-[0-9]{3} \|'` is the row count. **Do not write any of them down.**
- **A daily report's count is not a RATE** — the watcher only sees what is still broken at 06:50.
- **An image tag quoted from a live `agent_definitions` row is stale by construction; the identical
  citation about a repo SEED file is permanent.** Read which artefact holds the tag.
- **Search live config by the name the ROWS use, not the name the ENTRY uses.**

## Owed / blocked

- **`rebuild-cascade.md`'s stored count — still owed, SEVENTH session running.** Still dirty
  (+3/−3, the other lane's REB-003 rewrite, restored from `stash@{0}` by this lane on 08-12 and
  verified per file); last commit still `7272d59d4` **2026-07-27**. Same-file-blocked: a pathspec
  commit would take it as a passenger. It is the drift check's only finding.
  ⚠ **Do not read a clean `git status` as resolution.** "Clean" has two causes on this tree — the
  other lane committed, or their work was swept out — and only the second leaves the blocker in
  place while looking resolved. **Assert the positive fact:** `git log --oneline -1 -- <path>`
  newer than 07-27 means it landed; still older means the work has gone somewhere and
  `git stash list` is the next question. (WRONG_CALLS 2026-08-12, this lane. Do NOT grow
  `KNOWN_STORED_COUNTS` to silence the finding.)
- **Nothing else is owed.** No dispatch in flight, no verdict awaited, no half-applied change.
- Stray `register/model-infrastructure.md.tmp_check` — still untracked, still another lane's.

## Things deliberately NOT done

- **No SUMMARY.** Five-headings test: the register's own state is materially what
  `SUMMARY_where_we_are_2026-08-12.md` says. The stash incident and the delivery repair are real
  work but they happened *around* the register, not to it. **The next inflection is uptake** —
  whether anyone acts on `DOC-077`/`DOC-078` — which is item 2 at the top of this doc.
- **The 881 repairable citations and the 4 dead ones**: not rewritten, by design (above).
- **The 13 sha-less entries were not "fixed"** by looking up their commits — the authoring rule is
  the real fix, and it is item 1.
- **No `--update-ratchet` run**; the coverage report is quiet.
- **Nothing pushed** by this lane, and no local upstream configured. Not this lane's call.
