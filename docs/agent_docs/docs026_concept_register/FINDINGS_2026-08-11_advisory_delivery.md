# FINDINGS — the pre-commit advisory reaches nobody about half the time, 2026-08-11

*`HANDOFF_2026-08-10b_continue_here.md` opens with a READ FIRST banner: OPP-006
fired and was ignored within three hours of shipping, and — the useful part —
**"the honest question to answer first is which of these it was"**: the session
never saw the output, or saw it and judged the row could wait. *"Those have
opposite fixes (delivery vs. enforcement), and nothing recorded so far
distinguishes them."* Something does distinguish them. This answers it, and the
answer generalises well past this lane.*

**The answer: the session never saw it.** Its `git commit` command ended
`2>&1 | tail -8`, and the stdout recorded in that session's transcript is
**exactly 8 lines long**. The pre-commit hook prints the advisory FIRST; git
prints its `[branch sha] subject` summary LAST; `tail -N` is a head-cutter.

**And it is not one session's slip.** Across every commit made through the Bash
tool since the commit-scope report shipped on 2026-07-18:

| multi-file commits in this repo, made through the tool | 2,669 |
|---|---|
| advisory block **delivered** to the session | **1,470 (55%)** |
| **not delivered** | **1,199 (45%)** |
| of those misses, cut by the session's own `\| tail` | **1,137 (95%)** |
| distinct sessions that suppressed at least one | **258** |

*Every figure here is `scripts/advisory-delivery-sweep.py` at 2026-08-11 ~19:30 —
run it rather than quoting these, because the corpus grows by the hour and the
whole point of the fix is to move the top two rows.*

---

## What was actually suppressed

Everything `.githooks/pre-commit` prints, which is not a small list:

- **the commit-scope report** — the estate's main defence against committing
  another session's work, and the thing CLAUDE.md's whole "commit per task"
  section leans on ("Read it: any file listed that is not part of your task
  belongs to another session");
- **every `pattern-check.py` finding** — all 17 checks, including OPP-003,
  OPP-004 and this lane's own OPP-006;
- **the architecture signal** — the mechanical RFC trigger test, which exists
  because RFC_001 was written after its code was already in production.

The commit-msg hook's output survives, because it prints later. That asymmetry is
why the estate reads as "hooks are working": the council-trailer nudge arrives, so
the machinery looks alive.

## Why it is systematic rather than sloppy

Nobody is ignoring anything. The output ordering makes the pipe a filter that
removes exactly the part that warns you and keeps exactly the part you asked for:

```
  ── commit scope: 5 files across 2 area(s) ──      ← pre-commit  (cut)
  ── pattern check: 2 things worth a look ──        ← pre-commit  (cut)
  3 platform-code file(s) staged with a trailer…    ← commit-msg  (kept)
  [087_towards_multiple_domains 5c7b115c5] fix(113)…← git         (kept)
   5 files changed, 375 insertions(+), 41 deletions ← git         (kept)
```

A session pipes to `tail` for a good reason — it wants the sha and the file count,
not the noise. The advisory is upstream of the thing it wants.

## The two controls, because a 45% figure with no control is a correlation

**Control 1 — delivery tracks the pipe width, and could have come out otherwise.**
If `| tail` were incidental, N would be unrelated to whether the block arrived.
It is not:

| `tail -N` | among the **misses** | among commits that delivered **despite** a pipe |
|---|---|---|
| N ≤ 8 | **1,031** | 130 |
| N > 8 | 106 | **465** |

Median N among misses: **5**; the four commonest are `-4`, `-6`, `-3`, `-5`. The
mechanism predicts a crossover and the data has one. The sweep prints this control
on every run, and if the separation ever disappears the cause claimed here is not
the one operating.

**Control 2 — arithmetic no content can fake.** For **666** of the misses the
recorded stdout is *exactly* N lines long for a `tail -N` command. Output existed
and was cut. (Where stdout is *shorter* than N, nothing was cut — those are
counted as misses only if the block was genuinely due, and they are the residue
this document does not claim to explain.)

**Scope check on the denominator.** Only multi-file commits are counted, because
`commit-scope-report.sh` deliberately exits silently for a single-file commit
(`[ "$n" -le 1 ] && exit 0`) — counting those would have manufactured ~2,600 fake
misses. Commits are also required to resolve in *this* repo: shas from other repos
entirely (the auto-memory git dir, scratch repos) are excluded, because no hook is
installed there and a miss means nothing. **That filter was not foresight** — the
first pass counted them, and 495 commits where nothing had been cut sat in the
result looking like unexplained misses. Chasing them is what found the foreign
shas. The residue after the filter is 62, and this document does not claim to
explain those.

## The wrong turn, recorded because it was cheap and instructive

My first hypothesis was **structural and wrong**: that `git commit <pathspec>` —
the form CLAUDE.md *mandates* — leaves the pathspec'd files out of the index, so
every check reading `git diff --cached` would be blind to precisely the commits the
house style produces. It is a tidy theory, it explains the symptom, and it is
false.

**Refuted in one scratch repo in about a minute.** Git builds a temporary index
for a partial commit and points the hook at it, so `git diff --cached` inside
pre-commit is faithful in all four shapes tested: bare commit after `add`;
pathspec commit of an unstaged modification; pathspec commit mixing `add`ed new
files with unstaged modifications (the exact shape of `5c7b115c5`); and a pathspec
commit naming a file while a *different* file sits staged.

The lesson is not "test your hypotheses" — it is that **the symptom was
consistent with two mechanisms and I reached for the interesting one.** The dull
one was correct. What separated them was reading the actual command that produced
the commit, which took one grep of the transcript and should have come first.

## Where the evidence came from, and the substitution declared

CLAUDE.md's owner ruling of 2026-07-31 says a cross-cutting root-cause claim is
not filed until it has been through the `090` diagnosis loop, **or** the filing
session states plainly why it substituted equivalent first-hand verification.
Stating it: **the loop could not have run this.** It reads the repo and the live
database; this defect lives in neither. The evidence is the harness's own session
transcripts (`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`, 1.4 GB,
573 files), and the verification is first-hand on all four legs — the offending
command read back verbatim, the population swept exhaustively rather than sampled,
two independent controls, and a scratch-repo refutation of the competing
mechanism. A `090` run would have had to take my word for every one of those.

## The fix, and what it deliberately is not

**OPP-007 — `scripts/commit-advisory-postuse.py`**, a `PostToolUse` hook on `Bash`.
When a Bash call contains `git commit` and its output carries git's summary line,
it re-runs `commit-scope-report.sh --commit <sha>` and `pattern-check.py --commit
<sha>` and delivers the result through `hookSpecificOutput.additionalContext` —
**out of band, so no pipe in the session's own command can touch it.**

- **No git hook can fix this.** `post-commit` output also lands *before* git's
  summary (verified in a scratch repo), so the same `tail -3` eats it too.
- **Delivery path is the load-bearing detail**: `additionalContext` on **stdout**,
  exit 0. A `PostToolUse` hook that writes to **stderr** and exits 0 reaches
  neither the model nor the transcript — `scripts/memory-index.py` was wired that
  way and was mute for six days.
- **`commit-scope-report.sh` gained a `--commit <sha>` mode** to make this
  possible (by then the index is clean), mirroring the mode `pattern-check.py`
  already had. Staged mode is untouched; both were re-controlled after the change.
- **It is not teeth.** `PostToolUse` runs after the tool: the commit already
  exists. This buys visibility one moment later, not enforcement. **The case for
  making OPP-006 blocking is not strengthened by any of this — it is weakened**,
  because the evidence that looked like "the gate is being ignored" was an
  artefact of a pipe. `pattern-check.py`'s own argument against blocking on a
  shared tree (a false positive that blocks is a fleet-wide outage; a check that
  blocks on a bad day gets disabled for ever) stands entirely untouched.

## What this does not settle

- **The residue.** 62 misses had no `tail` pipe and are unexplained here.
- **Whether being told changes anything.** Delivery is necessary, not sufficient:
  a session can read the block and still not act. The honest test is OPP-007's
  verify-later — if the delivery rate rises and the daily watcher's missing-row
  count does *not* fall, then delivery was never the binding constraint and the
  enforcement question reopens on real evidence rather than on this artefact.
- **The other three staleness signals** from `FINDINGS_2026-08-10` — version lag
  (80 entries 50+ versions behind), unresolvable citations (96), moved bug
  references (156) — are untouched and remain this lane's open work.

## Reproducing the measurement

```bash
scripts/advisory-delivery-sweep.py                    # the full window
scripts/advisory-delivery-sweep.py --since 2026-08-12 # after the fix, for the verify-later
```

~5s over 1.4 GB of transcripts. It reads only local transcript files and the local
git object store; nothing touches the cluster. `RUNBOOK_concept_register.md` §B11
carries the gotchas.
