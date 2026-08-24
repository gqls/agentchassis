# HANDOFF — orchestration status retention (the `566`/`589` lane) — 2026-08-24

**STATUS: this lane is COMPLETE and can be CLOSED.** Everything it set out to do is applied,
verified at the artefact and committed. Nothing is mid-flight, nothing is owed to another lane,
and there is no dispatch in the air. This document exists so a future session can (a) confirm
that quickly, (b) find the material, and (c) know what is deliberately NOT done and who owns it.

**Do not create the standing five for this lane.** The work was executed as contributions INTO
existing lanes' documents, which is where the material actually lives and where a reader will
look — see "Where the material lives" below. A parallel account here would be the drifting second
account CLAUDE.md warns about.

---

## 1. What this lane was

It began as a one-line request: *pick up the remainder of `bugs_open/354`.* That remainder was a
finished-but-orphaned migration, `566`, written by the `354` lane on 2026-08-22 and left
**untracked and unapplied** when that session ended. Two other lanes (`bugs_open/358`,
`bugs_open/307 [abdc1e]`) found it, independently declined to adopt it, and recorded it rather
than let it vanish (`6dd0e01a6`). The owner then directed this session to take it on.

It grew one step, on evidence rather than ambition: `566`'s own council review raised a hazard
that `566` had created, and the owner asked for that to be closed too (`589`).

## 2. What shipped — all applied, all verified

### `566` — `database-cleanup` reaps EVERY terminal status, not a literal pair
- **File:** `docs/agent_docs/sql_for_agents/566_database_cleanup_reaps_every_terminal_status.sql` (+ `_ROLLBACK`)
- **Applied** 2026-08-23 17:46Z · **commit `ccc851a42`** · council `9d23ccd9-c16c-422d-8bf9-7b60e8b52795` APPROVED
- **The defect:** arm 3 deleted `WHERE status IN ('COMPLETED','FAILED')` — a literal — while arm 4
  skipped everything `is_terminal`. A terminal status named by neither arm was reaped by NOTHING.
- **The damage it was doing:** `CANCELLED` had been in that position since migration `466` — 24
  rows, oldest 2026-07-19, **35 days** against a 24-hour retention norm.
- **Proof it worked:** the 2026-08-23 18:22Z sweep took those 24 rows to **zero**, with **4,789**
  rows younger than 24h untouched as the control (a fix that reaped everything would pass the
  first check and be worse than the bug).

### `589` — a status cannot be both `is_terminal` and `is_pausable`
- **File:** `docs/agent_docs/sql_for_agents/589_a_status_cannot_be_both_terminal_and_pausable.sql` (+ `_ROLLBACK`)
- **Applied** 2026-08-24 (by the owner — see §5) · **commits `9e0b0daa9`** (written) and **`9d060db04`** (live + docs) · council `fbf9bcc2-e3e8-4fca-8036-8a63ca2763ec` APPROVED, 9 seats
- **The defect `566` created:** once arm 3 read `WHERE is_terminal`, a row marked BOTH terminal and
  pausable was **deleted by arm 3 while arm 4 spared it** — arm 3 never reads `is_pausable` at all,
  so the "never reap this" protection was silently void and the destructive arm won. The damage
  would be live human-in-the-loop orchestrations deleted 24h after `updated_at`, looking exactly
  like ordinary reaping.
- **Two halves, deliberately:** a `CHECK (NOT (is_terminal AND is_pausable))` stops the bad row
  being WRITTEN; arm 3's new `AND NOT is_pausable` stops it causing DAMAGE if it exists anyway.
  The second is unreachable while the first holds — **that is the point, not dead code**: the
  constraint is one `ALTER TABLE` from being dropped, and 22 characters is what still protects
  paused work that day.
- **Inert on live data by design:** 7 statuses, 3 terminal, 0 pausable, 0 both. A guard against a
  future write, not a repair.

### `DBI-026` — the seam, registered
- `docs/agent_docs/docs026_concept_register/register/database-and-infrastructure.md` + its index row
- Filed because `589` ALTERS what a shared mechanism guarantees, which the 2026-07-28 owner ruling
  condition (2) requires registered in the same commit that ships it. Carries the landmine and
  **two open review questions with their triggers** (§6). No DBI drift (entry ↔ index checked).

## 3. How to confirm all of that is still true

One query. Every value below was true at 2026-08-24 18:32Z, after a fresh chassis roll:

```sql
SELECT (SELECT count(*) FROM pg_constraint
         WHERE conname='chk_status_not_terminal_and_pausable')            AS constraint_live,   -- 1
       pre_query LIKE '%WHERE is_terminal AND NOT is_pausable)%'          AS arm3_guarded,      -- t
       pre_query LIKE '%''COMPLETED'', ''FAILED''%'                       AS literal_regressed, -- f
       (SELECT count(*) FROM schema_migrations
         WHERE filename LIKE '566%' OR filename LIKE '589%')              AS ledger_rows,       -- 2
       last_completed_at >= last_triggered_at                             AS sweep_not_erroring -- t
  FROM scheduled_tasks WHERE name='database-cleanup';
```

`sweep_not_erroring` is the load-bearing one: the whole `pre_query` is **one statement**, so a
fault stops ALL SIX arms, not just the orchestration ones. Three consecutive clean hourly runs
were observed on `589`'s query (16:32, 17:32, 18:32Z).

**⚠ `updated_at` on that row is NOT an edit signal** — the scheduler writes to the row every run,
so it bumps hourly with no content change. Only `md5(pre_query)` tells you whether the text moved.

## 4. Where the material lives (paths, because names are ambiguous here)

| what | where |
|---|---|
| the mechanism, as a register entry | `docs/agent_docs/docs026_concept_register/register/database-and-infrastructure.md` → `DBI-026` |
| the trap, prospective, for anyone touching the table | `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` → "Setting `is_terminal` … now ARMS a 24-hour DELETE" |
| how to add a status, with the corrected claims | `docs/agent_docs/docs024_key_docs_latest/orchestration_status_lifecycle/RUNBOOK_orchestration_status_lifecycle.md` |
| the originating bug (NOT this lane's — see §5) | `bugs_open/354_HANDOFF_2026-08-22_a_workflow_that_ends_at_its_error_terminal_is_recorded_COMPLETED_with_error_NULL.md` |
| the lane that recorded the orphan | `docs/agent_docs/docs024_key_docs_latest/bugfix_358_unread_finding_codes/NOTES_unread_finding_codes.md` |
| the note handed to the register lane | `docs/agent_docs/docs026_concept_register/RUNNING_NOTES_concept_register.md` (2026-08-24 entries) |
| submission-practice lesson | `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/RUNBOOK_council_gate.md` |

**Commits, in order:** `ccc851a42` · `27a57703c` · `bc9f91da0` · `15aea43a1` · `5d3384ff5` ·
`e61e901da` · `9e0b0daa9` · `75eb788ec` · `9d060db04`.

## 5. What is NOT done, and who owns it — read this before "finishing" anything

**`bugs_open/354` itself is still OPEN and is NOT this lane's to close.** It is owned by the
`bugs_open/307` lane. Its actual defect — a workflow ending at its error terminal recorded
`COMPLETED` with `error` NULL — is **unfixed**; its candidates 1 (a distinct terminal status,
architecture-scope) and 2 (populate the `error` column) are both unshipped. This lane only
**removed a cost** from candidate 1: it no longer has to widen the cleanup arm itself, because
`566` did that on its own account. That contribution is recorded inside `354` under a dated
UPDATE, explicitly saying it does not change §5's ordering. **Do not read this lane's completion
as progress on 354's root cause.**

> **⚠ AND THERE IS A LIVE HAZARD SITTING ON 354's LANE — read this before you touch
> `platform/orchestration/coordinator.go`.** Flagged by the `bugs_open/243` lane on 2026-08-24
> (`eb85fbbc7`) and **re-verified first-hand by this session, not taken on trust**:
>
> ```
> git status --porcelain platform/orchestration/coordinator.go          # -> ' M' (dirty)
> git status --porcelain platform/orchestration/error_route_completion*.go  # -> '??' both
> git show HEAD:platform/orchestration/coordinator.go | grep -c errorRouteTermination  # -> 0
> grep -c errorRouteTermination platform/orchestration/coordinator.go                  # -> 1
> ```
>
> The working tree calls `errorRouteTermination`; its callee lives in two **untracked** files.
> So **any session that commits `coordinator.go` breaks HEAD** with `undefined:
> errorRouteTermination` — and cannot see it coming, because a pathspec commit takes the whole
> working-tree file (that is what it is for), the untracked callee would not ride along, and no
> commit-scope report can show a same-file passenger. This lane touched none of those files.
>
> It is already costing a third party: `243` is holding a finished ~20-line addition to
> `routeToErrorStep` out of the tree for exactly this reason, with its migration `588` held
> behind it. **Do not "tidy" this by committing `coordinator.go`.** The two ways out are the
> 243 lane's: add all three files together, or drop the call and leave the new files untracked
> (an untracked file alone breaks nothing).

**`DBI-014`'s drifted figures are the concept-register lane's repair, not this one's.** Measured
2026-08-24: `awaited_requests 7 days` and the `orchestration_requests` CASCADE still hold;
`agent_error_log 14/30 days` and `orchestrations 7 days/24h stuck` do not (the live sweep has **no
7-day interval at all**). Two lanes superseded those facts in two weeks — migration `567` for the
log arm, `566` for the orchestration arms — neither knowing the entry existed. The note to that
lane also flags a cross-file contradiction their two staleness reports structurally cannot see:
`DBI-014` asserts the "always-return-a-row" premise that their own `SCH-007` corrected on
2026-08-17. Offered to them as a candidate worked case for their open "are the reports actionable?"
question; **it is theirs to take or decline.**

**The apply is gated and always will be.** `--apply` is refused by the auto-mode safety classifier;
it needs the owner to run it, naming the target. The command shape that works, scoped so it cannot
sweep other lanes' pending migrations:

```bash
D=$(mktemp -d) \
  && cp docs/agent_docs/sql_for_agents/<file>.sql "$D"/ \
  && MIGRATIONS_DIR="$D" ./scripts/migration/run-migrations.sh --apply; rm -rf "$D"
```

This is better than the `psql -f` recipe in the lifecycle runbook because the runner's **probe**
executes the file verbatim in a doomed transaction first — a stale md5 guard, a lost arm or a query
that no longer parses is caught before anything is written — and it records the ledger row itself,
with no separate `--record-only` to forget.

## 6. Open questions, with their triggers — not owed now

- **(a) Could a status ever legitimately need both flags?** `589` assumes not — that ending and
  waiting-for-ever are contradictory. Nobody has tested that against a real pause-for-human design.
  **Trigger:** the first time someone wants both.
- **(b) Is 24h the right ceiling for `AWAITING_RESPONSES`?** It is `is_pausable = false`, so arm 4
  deletes it at 24h, while its own `notes` say *"The reaper spares it while that map is non-empty"*
  — that sentence is about the separate `stale-orchestration-reaper` task, so the two are not in
  conflict, but the ceiling was **not** examined. This is the likeliest place a real pausable status
  first appears. **Trigger:** any complaint about awaiting orchestrations disappearing.
- **(c) Should this predicate still live in a string at all?** Raised by the council `guardian` seat
  (low severity). `589` was the **third** hand-edit to `database-cleanup`'s single `pre_query` blob
  in three days (`566` repointed arm 3, the `358` lane rewrote arm 1's comments, `589` narrowed arm
  3 again). Each was anchor-guarded and individually sound; the pattern is not. **Trigger: a fourth
  edit, or any edit that must touch two arms at once** — then move it into a view or function.

## 7. Traps this lane hit, so the next session does not

- **An md5 guard must be computed IN the database.** `length()` counts CHARACTERS while `md5()`
  hashes BYTES, and this row holds a multi-byte character, so a locally-hashed `psql` dump differs
  by 3 bytes. A locally-derived after-md5 would have aborted the apply at the migration's own
  byte-exact assertion. (Known family: `LANDMINES.md`, "`length()` on stored HTML is CHARACTERS".)
- **A guard observed only passing cannot be told from one that cannot fire.** `589`'s verify block
  deliberately violates its own constraint and requires the violation to be caught — and that test
  was itself proven to discriminate by running it against the live table BEFORE the constraint
  existed, where it correctly failed. Its control (a LEGAL write must still succeed) is what stops
  it passing merely because every write to the table fails.
- **A zero needs a demand control.** The blast-radius check for `566` returned 0 across five
  referencing tables. That only meant something because the same join key returned **4,069** rows
  against `COMPLETED`, and because four of those tables turned out to be EMPTY — so the zero was
  honest rather than a broken query.
- **An objection naming one file is naming a CATEGORY.** `editquality` objected on two consecutive
  rounds that the rationale claimed a file the edits list did not show — round 1 the `_ROLLBACK`,
  round 2 the register entry. Answering the first instance specifically is what earned the second.
  Now a rule with a mechanical pre-check in the council runbook.
- **A landmine can be falsified by your own change, hours after someone writes it.** The
  `orchestration_states` retention entry was written on 2026-08-23 and `566` invalidated its
  headline at 18:22Z the same day. Corrected in place, with the general form re-pointed at the
  `is_pausable` class — which is empty today, so the clean two-day window is **not** permanent.
- **The landmine verifier returns `NEEDS_HUMAN_REVIEW` on this entry and always will.** Its index is
  Go-only (~8,400 symbols); this entry's footprint is SQL, markdown and DB tables it structurally
  cannot reach. That is a scope limit, not a defect in the entry — do not "fix" it.

## 8. If you are picking this up cold

Read §3 first and run its query. If every value matches, **this lane needs nothing from you** —
close it, and go and read §5 to see whether the work you actually came for belongs to `bugs_open/354`
(the 307 lane) or to the concept-register lane instead.
