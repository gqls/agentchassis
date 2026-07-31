# PLAN — bugs_open/135: a floor on the code_symbols reconciliation prune

**Started** 2026-07-31 by the "bugfix 11" session. **Case file:**
`bugs_open/135_HANDOFF_2026-07-28_code_symbols_prune_has_no_floor_and_can_delete_a_working_index.md`.

## Why this bug, and how it was picked

The instruction was "the next bug in `bugs_open/` that isn't being worked on in
another thread". That is harder than it sounds on this tree: `scripts/who-owns.py`
returns *OWNED or recently active* for almost everything, because its window is 14
days and this repo takes ~1,500 commits a week. So the discriminator used was
**last commit touching the bug file** plus a read of the tree for in-flight WIP:

- 033/066/071/072/084/085/093/096 — all have an owning workstream directory with
  commits this week. Left alone.
- 119 (malformed-JSON council seat) — same code as `bugs_open/138`, whose lane
  committed today. Collision risk, skipped.
- 143/152/155 — `derive_card_asset` / `deploy_image_asset`. The tree shows
  `asset_lock_guard.go` **untracked** and four related actions dirty: a lane is
  mid-fix right now. Skipped (this is the case memory warns about — who-owns reads
  commits, so a session mid-fix is invisible).
- **135 — taken.** Filed 2026-07-28, last touched 2026-07-28, and *explicitly*
  split out of the markdown-indexing plan so that a different thread could take
  it: "The defect is pre-existing and independent … so it belongs here, not as a
  rider on someone else's change."

## Is the bug still valid? Yes — read 2026-07-31

`code_symbols_actions.go` still ends every run with

```sql
DELETE FROM code_symbols WHERE repo = $1 AND commit_sha IS DISTINCT FROM $2
```

guarded by nothing but `if commitSHA != ""`. Nothing checks that the run saw the
repo. Live population, same day: one repo (`gqls/agentchassis`), 4,992 rows, five
kinds (func 3048, method 1025, struct 857, interface 33, alias 29), **all at one
commit** — so the index is healthy today, and arming a floor now costs nothing.

## The design, and where it came from

**The design is not mine.** Candidates 1–3 of the case file are what six council
rounds on corr `7ba5b8c4` produced before the scope was split, and the case file
says so explicitly: "recorded here so the next thread starts from the end of that
argument rather than the beginning." So this plan implements them rather than
re-deriving them, and the only genuine design decisions left were (a) the
granularity of detection, (b) the remedy when detection fires, and (c) how far to
generalise.

### Decision 1 — cohorts, not one total (detection granularity)

A single whole-corpus ratio is not sensitive enough. A run that re-confirms 95% of
rows can still have dropped **100% of one class**, and the total hides it
completely. So the caller partitions into cohorts:

- **one per symbol kind.** This is also what protects a class this code has never
  heard of: a Go-only run cannot delete a future markdown corpus, because that
  cohort reads 0% confirmed and refuses. The markdown plan this bug was split out
  of is thereby protected by a guard that does not know it exists.
- **one in distinct paths** — a different *unit*, deliberately. Rows measure how
  much the run WROTE; paths measure how much of the repo it SAW, and "saw" is the
  property that actually fails when a tarball arrives truncated. This is the
  case file's candidate 4 ("a whole-repo signal … not costed") done with data
  already in the table, rather than with the analyser's `file_count`, which has
  nowhere to be stored for comparison and would need a schema change.

### Decision 2 — the refusal is all-or-nothing (the remedy)

Detection is per cohort; the *response* is to skip the whole prune. Considered and
rejected: pruning the kinds that passed and retaining the ones that failed.

Rejected because **a refused prune is self-healing and a wrong delete is not.**
The DELETE is defined against the *current commit*, not against *this run*, so the
next healthy run deletes everything a refused run retained. The cost of refusing
too much is therefore one cycle of staleness; the cost of deleting too much is the
index. A half-firing guard is also harder to reason about for a benefit that
property mostly erases.

### Decision 3 — fail closed on an unmeasurable floor

If the cohort query errors, the prune does not run. The measurement is the only
thing between a partial run and a deleted index. (Contrast `bugs_open/143`'s
finding about `assets.status`: a guard conditioned on something unconstrained
fails *open*, which is how a guard ends up in a guard's costume.)

### Decision 4 — how far to generalise, and where to stop

`grep -rn 'DELETE FROM' --include=*.go` finds three other live call sites with the
identical destructive shape: `populate_nav_tables_action.go:147` (whole-site nav
wipe-and-rebuild), `site_db_actions.go:1474` (`link_registry` per page),
`save_page_sections_action.go:532` (agent-writable `page_components` per page).

So the **rule** is generalised — its own file, pure functions of counts, no SQL,
no knowledge of any one table — while the **measurement** stays per-table, because
that is the half that genuinely differs. The three other sites are **not**
converted here: each needs its own cohorts *measured* rather than assumed, and two
of them are another lane's live territory this week. They are named in the new
file's header and in the register entry's open-review-question, which is the point
of putting the rule in its own file at all: the next thread starts from the end of
this argument.

> This is a **platform seam** by CLAUDE.md's definition (a shared mechanism whose
> blast radius is "every caller"), so: registered in the concept register
> (**CTXA-025**) in the same commit that ships it, and submitted to the council
> gate (corr `14239fa4-552f-4821-abaf-ea15ccee4ea5`) before the commit. Per the
> owner ruling of 2026-07-29 no ordering constraint is claimed — there is none;
> review here is after the fact by design, because HEAD is shared and any other
> session's build ships this commit.

### Decision 5 — close the read-side hole this fix opens (added mid-implementation)

Not in the case file, and it should have been. A refused prune **retains stale
rows on purpose**, and the read side's freshness banner reads exactly one row
(`ORDER BY updated_at DESC LIMIT 1`) and announces *its* commit. After a refusal
the newest row is this run's, so a part-stale index would be described as being at
the new commit. That is `bugs_closed/108`'s lie one layer along — and 108's fix
deliberately made the empty answer *more* confidently worded, so it would lie
harder.

So `codeIndexScope` gains a distinct-commit count (folded into the COUNT query it
already runs — no extra round trip) and a note that fires **only** when the corpus
spans more than one commit. Both readers of that helper get it.

Worth recording that the state was **already reachable** before this change, via
the pre-existing "no commit_sha → prune skipped" branch. So this closes a
pre-existing gap that the fix would have made more likely, rather than a gap the
fix invents.

## Phasing

1. ✅ Rule + tests (`prune_floor.go`, `prune_floor_test.go`) — pure, no DB.
2. ✅ Wire at the one call site: cohort SQL, fail-closed, durable refusal note,
   reporting (`prune_status`, `files_analysed`, `prune_cohorts`).
3. ✅ Read-side note + test.
4. ✅ Council submission + register entry.
5. ✅ Commit, build, roll.
6. ⏳ **Induce the refusal live** and watch it fire. A green run over a healthy
   repo proves only that the guard is inert — the case file says exactly this, and
   it is the difference between "shipped" and "works".
7. ⏳ Close to `bugs_closed/` once (6) has happened.
