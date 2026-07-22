# HANDOFF — applied-but-unrecorded migrations block the runner (ledger-omission trap)

> ## STATUS 2026-07-21 — STILL OPEN, but the tooling shipped and is LIVE
>
> **Cold start:** read this banner, then jump to **"FIXED 2026-07-20 — the tooling
> landed"** near the bottom for what changed and how it was verified. The
> 2026-07-16 sections below are the original diagnosis and remain accurate history.
>
> **What is done and LIVE** (commits `a51333fd7` + `ed1f70396`; shell only, so
> live on commit — no image roll): `run-migrations.sh` now has
> `--record-only <file> --note "<why>"` to register an out-of-band apply; a
> failure message that names the already-applied cause and prints the recovery
> command; sidecar (`_ROLLBACK`/`_VERIFY`) exclusion; a dry run cut from >120s to
> ~5s; a refusal when the DB is unreachable; and an advisory idempotency lint on
> pending files. This is **all three** non-rejected fix candidates — **(a)** and
> **(b)** and **(c)** — plus three defects **(e)(f)(g)** found while taking the
> baseline dry run.
>
> **Why it is still OPEN, not moved to `/bugs_closed/`** (the bar there is
> *fixed AND live AND not reproducible*). Both defects the trap needs are now
> *made-easy-to-avoid* — recording is one command (a), the replay cause is named
> (b), non-idempotent pending files are flagged (c) — but **none is enforced**,
> so the blocking symptom is still reproducible:
> 1. **The runner still HALTS on an already-applied-but-unrecorded migration.**
>    Fix (b) made that halt *informative* and gave a one-command recovery; fix (c)
>    warns *before* you apply a non-idempotent file — but neither makes the runner
>    auto-detect-and-skip (candidate (d), auto-record, was deliberately rejected as
>    unsafe: a 23505 can also be genuinely-wrong SQL, and auto-recording would mark
>    broken SQL as applied — worse than a replay). Recording and heeding the
>    warning are both manual, so a thread that ignores them can still gate the
>    queue. The *diagnosis cost* (the 3-day misread) is fixed; the *block itself*
>    is mitigated, not eliminated.
> 2. ~~**The `--record-only` INSERT path is UNEXERCISED against the production
>    ledger.**~~ **CLOSED 2026-07-22.** The path is now in active fleet use:
>    8 `applied_by='record-only'` rows exist (183 first, on 07-21; then 177/178/
>    179/187/189/190/193 on 07-22). Verified idempotent under real contention — a
>    concurrent session recorded 193 in the seconds before this thread tried, and
>    the `ON CONFLICT DO NOTHING` no-op fired cleanly ("already recorded — nothing
>    to do"). See **RECURRENCE 2026-07-22** below.
>
> **What IS eliminated:** near-miss (e) — the runner treating 180's `_ROLLBACK.sql`
> as a pending migration and reverting bug 024 — is closed and live.
>
> **Next actions for a resuming chat**, in priority order:
> - (nothing forces action — this is a mitigated landmine, not an outage.)
> - **Residual 2 is DONE** (see above). The one remaining lever before 007 is
>   arguably closable is residual 1 (the runner still halts). The 2026-07-22
>   recurrence hardened the case for a *mechanism* and proposes a **safe** design
>   for the first time — read that section before deciding to build or to leave it.
> - **Run the dry run at the START of any session that will touch migrations** —
>   the 2026-07-22 recurrence found the queue lying about SEVEN files at once, and
>   the set churns every few minutes under concurrent load. "Pending" is now more
>   often "applied-and-unrecorded" than "genuinely pending" — verify each before
>   believing either.
> - If you want the runner to stop halting at all (residual 1): the honest options
>   are (i) leave it — "die loudly with a fix in the message" is a defensible
>   design, or (ii) reconsider auto-detect (skip-if-already-applied). Note (ii)
>   changes the runner from "die loudly" to "carry on", which the 2026-07-16
>   analysis was wary of *for the auto-record variant*; a skip-only variant that
>   still refuses to *record* is a middle path worth designing before building.
> - Do **not** `--apply` casually: pending files are usually another thread's, and
>   applying someone else's migration can violate an image-first ordering.

**Created 2026-07-16 from the travelling-docs workstream** (its HANDOFF T23 has the discovery
blow-by-blow). The INSTANCE is already resolved; what needs attention is the SYSTEM that let it
happen and will let it happen again. **Not an outage.** Runner:
`scripts/migration/run-migrations.sh`; ledger `schema_migrations` (keyed on `filename`); migrations
home `docs/agent_docs/sql_for_agents/`; DB access
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

**Severity: process landmine, fleet-wide.** For ~3 days every workstream believed migrations were
"blocked by someone else's broken SQL". Three real migrations (157–159) had to be applied out of
band, and the block was mis-attributed in three separate workstreams' docs.

---

## What happened (the resolved instance)

The empty-sections/loop-integrity workstream applied its SQL **151–156 by hand** (they predate its
adoption of the runner) and **never inserted `schema_migrations` rows**. The runner therefore saw
them as pending, replayed `151_gripper_spec_sheet_component.sql`, hit its **duplicate
content_components insert** (the row already existed — the file had already run), and halted —
and because the runner stops at the first failure, **everything numbered after it was gated** for
every workstream.

The failure was misread by everyone, in three different ways:
- the owning workstream called the numbering collision "cosmetic only" (its HANDOFF_2026-07-16 §7);
- the travelling-docs workstream recorded "gripper-151 FAILS on a duplicate and blocks the runner —
  being fixed in a separate chat" (i.e. assumed broken SQL, waited);
- nobody owned the actual defect, because an applied-but-unrecorded migration **fails wearing the
  look of broken SQL with the author's name on it**.

**Resolution applied (2026-07-16):** verified every file's artifacts live in the DB
(gripper-spec-sheet component ×1; 5 gripper products; the detail-page slot layout; the
`gripper-spec-sheet` `site_plan_sections` row; `section_source_drift` present in
**completeness-discovery-agent**'s checks — note: NOT design-discovery-agent, a naive verification
looks in the wrong agent and reports "missing"; active `product-spec-refresher` agent), then
backfilled six ledger rows with `applied_by='ledger-backfill'` and a note citing the owning handoff.
Runner dry-run now reports **"Up to date — no pending migrations."**

## The two defects that must BOTH be present for the trap to spring

1. **A migration applied without a ledger row.** The runner records what IT applies; an
   out-of-band apply (psql -f by hand) records nothing, and there is no tooling support for doing
   so — today it takes a hand-written `INSERT INTO schema_migrations …` (the travelling-docs
   workstream did this for its 157/158/159, which is why those never blocked anyone).
2. **A non-idempotent migration.** The shop convention (stated in the runner's own header comment)
   is "every migration carries its own guard DO block" — a guarded file replays as a no-op and the
   trap never springs. 151's `INSERT INTO content_components` had **no guard and no ON CONFLICT**,
   so the replay errored instead of skipping.

Fixing either defect prevents the block. Fixing both is the robust position.

## Fix candidates (the "maybe fixing the migration code" part)

**(a) `--record-only <file>` flag on run-migrations.sh — RECOMMENDED, smallest.**
One command to register an out-of-band apply: computes the md5, inserts the ledger row with
`applied_by='record-only'` and a required free-text note (e.g. "applied by hand 2026-07-15,
artifacts verified"). Removes the excuse for defect 1 — today the manual INSERT is fiddly enough
that people skip it.

**(b) A better failure message — cheap, high value.**
When a file fails, the runner currently prints `!! $f FAILED — stopping.` Add the hint that would
have saved three days:
```
If this file may have ALREADY been applied out of band (duplicate-key errors are the classic
sign), verify its artifacts in the DB and record it with: run-migrations.sh --record-only <file>
```
Duplicate-key errors (SQLSTATE 23505) on replay are precisely the signature of an
applied-but-unrecorded file.

**(c) A lint for unguarded migrations — optional, preventive.**
Dry-run mode could warn on pending files containing a bare `INSERT INTO` with no `ON CONFLICT`,
no `WHERE NOT EXISTS`, and no guard `DO $$` block — the convention exists; nothing checks it.
(Warning only: some inserts are legitimately meant to fail loudly on duplicates.)

**(d) NOT recommended: auto-record on duplicate-key failure.** Tempting, but a 23505 can also mean
a genuinely wrong migration colliding with unrelated data — auto-recording would mark broken SQL
as applied. Keep the human in the verify step; make the tooling make that step easy (a) + obvious (b).

## The process rule (already banked in both workstreams' docs, restated here)

> **Whoever applies a migration records it.** The runner does this automatically; anyone applying
> out of band (`psql -f`) must insert the `schema_migrations` row themselves, immediately, in the
> same sitting. An applied-but-unrecorded migration becomes a runner roadblock that looks like
> someone else's broken SQL.

## Verify after fixing

1. `run-migrations.sh --record-only <file>` inserts a correct row (filename, md5, note) and a
   second invocation is a no-op (`ON CONFLICT DO NOTHING`).
2. Force a duplicate-key failure on a scratch file → the failure message shows the hint.
3. Dry run stays "Up to date" on the production ledger (nothing regressed).

## RECURRENCE 2026-07-20 — it sprang again, three files, three different workstreams

Found incidentally by a bugfix thread working `bugs_open/002`. The dry run reported
**3 pending** — `157_restore_gripper_spec_qualifiers.sql`,
`170_tool_list_card_image.sql`, `173_load_existing_pages_build_status.sql` — all three
**already applied and unrecorded**, by three unrelated threads. The process rule above
did not hold; ~4 days after it was written, the trap was fully re-armed.

**Verified live BEFORE backfilling** (the check is the load-bearing part — a ledger row
for a migration that never ran skips it *forever*, which is strictly worse than a
replay):

| file | artifact checked | result |
|---|---|---|
| 157 | `products.specifications->>'stroke'` for the two named grippers | `6 mm per jaw` / `10 mm per jaw`, voltage `24 V DC` — present |
| 170 | `content_components.tool-list.html_template` | contains `{{if .image}}`, retains `tl-card-icon` |
| 173 | `build-site-planner` `load_existing_pages` query | contains `p.build_status` and `adoption_locked` |

Backfilled with `applied_by='ledger-backfill'` and a note naming the owning handoff and
the exact check. Dry run now reports one genuinely-pending file
(`177_council_tolerate_truncation.sql`, another thread's — deliberately left for its
owner, since applying someone else's migration can violate an image-first ordering).

**What this recurrence actually tells us.** The process rule is necessary and is not
sufficient — it asks every thread to remember a second step, at the exact moment its
task feels finished, in a repo where threads routinely hand off mid-task. It failed
three times independently. That is not three careless threads; it is a rule doing a
job that wants a mechanism. It strengthens the case for fix candidate 2
(**`--record-only`**) and, more so, for the runner **detecting** an already-applied
migration rather than dying on it. Until one of those lands, expect this to recur, and
**run the dry run before believing the queue is clean** — the queue lies in the
optimistic direction (it reports work that is already done, so the danger is a replay,
not an omission).

### Numbering collision, 2026-07-19/20 — do NOT "tidy" it

Two `175_*` and two `176_*` files now exist (`175_experience_context_component_js.sql` +
`175_robot_hands_contact_plan_sections_fix.sql`; `176_experience_compose_length_and_quoting.sql`
+ `176_leopardess_aspect_generic_text_block_fix.sql`), from two threads picking the next
number concurrently. **This is harmless and must be left alone.** The ledger is keyed on
`filename`, not on the number, so all four are recorded and none replays; the runner
orders by `sort` and applies each once. **Renaming any of them to de-duplicate the number
would break its ledger row and make an applied migration look pending again** — i.e. it
would re-arm exactly the trap this file documents. The 2026-07-16 instance already
recorded a numbering collision being dismissed as "cosmetic only"; the correct reading is
that the number is cosmetic and the *filename* is the key.

## FIXED 2026-07-20 — the tooling landed (fix candidates a + b, plus two more)

`scripts/migration/run-migrations.sh` now carries the mechanism the recurrence
section above asked for. Shell only — no image roll, so this is live on commit.

**(a) `--record-only <file> --note "<why>"`** — registers an out-of-band apply.
Inserts `applied_by='record-only'` with a **required** free-text note; refuses a
file that does not exist, refuses a sidecar (below), and is a no-op on second
invocation. The note is mandatory precisely because the artifact check is the
load-bearing part and a flag with no note invites recording without checking.

**(b) The failure message now names the likely cause.** On any failed file the
runner prints that a duplicate key (23505) most likely means already-applied-
and-unrecorded rather than broken SQL, and prints the exact `--record-only`
command to copy, with the warning that recording an unrun migration skips it
for ever.

**(e) NEW — the runner would have applied ROLLBACK and VERIFY sidecars.** Found
while taking the baseline dry run for this fix. The candidate regex
`^[0-9]{3}_[A-Za-z0-9_]+\.sql$` matched `180_tool_improver_rerender_request_ROLLBACK.sql`
and `..._VERIFY.sql`, so both were listed as *pending migrations*. The ROLLBACK
strips the four keys 180 added (**re-opening bugs_open/024**) and ends with
`DELETE FROM schema_migrations WHERE filename = '180_tool_improver_rerender_request.sql'`.
Its own guard is no protection — the guard fires when 180 is *absent*, and 180
**is** applied (ledger row 2026-07-20 19:13), so the guard passes and the revert
proceeds. `sort` puts `.sql` before `_ROLLBACK.sql`, so on a directory where the
base migration was also pending it would apply and then immediately be undone in
the same run.

> **Precisely how live was it:** conditional, not certain. The runner stops at
> the first failure, so the ROLLBACK only executed if 177, 178 and 179 all
> succeeded first. It self-heals on a *subsequent* run (deleting 180's row makes
> 180 pending again, so it re-applies) — but in the window bugs_open/024 is
> genuinely re-opened, and it leaves a `doc_notes` row announcing the rollback
> plus a spurious snapshot. Recorded as a near-miss, not an outage.
>
> Fix: exclude `_[A-Z][A-Z0-9_]*\.sql$` from candidates and **list them** under
> a "Sidecars (hand-run only)" heading — reported, never silently dropped, for
> the same reason the odd-filename warning exists. The uppercase rule matches
> the live convention: of 176 files in the directory, the only two containing
> any uppercase letter are exactly these two sidecars.

**(f) A dry run took over 120 seconds; now 4.8.** It made one `kubectl exec`
round trip *per file* to test the ledger. The whole ledger is now fetched in a
single query. This is not cosmetic: the recurrence section's own advice is
"**run the dry run before believing the queue is clean**", and a two-minute
check is one that gets skipped.

**(g) An unreachable database no longer reads as an empty ledger.** Every psql
call swallowed stderr, so a failed `kubectl exec` returned "" — indistinguishable
from "the ledger table does not exist", which the runner treated as *nothing is
applied*. With `--apply` that meant replaying the entire history from 124. The
runner now probes `SELECT 1;` first and refuses to do anything if it fails.

**(c) An advisory idempotency lint on pending files** (commit `ed1f70396`, added
2026-07-21). The dry run and `--apply` both warn (stderr, never blocking) when a
pending file INSERTs into a non-log table while carrying no guard `DO` block, no
`ON CONFLICT` and no `WHERE NOT EXISTS` — the exact shape that errors on replay.
It doubles as an early warning for *this* trap: a non-idempotent pending file is
precisely the one that blows up if it turns out to be already-applied-and-
unrecorded, so the warning points at `--record-only`.

> **Calibrated against the live corpus, not the handoff's literal "any bare
> `INSERT`" spec.** Nearly every migration writes a `doc_notes` audit row, so a
> literal lint would fire on almost every file and be learned-ignored within a
> day. The allowlist — `doc_notes`, `doc_plans` (append-only logs) and
> `schema_migrations` (self-record, always `ON CONFLICT DO NOTHING`) — is the
> load-bearing refinement. Validated: it flags 151 (the original culprit,
> `INSERT INTO content_components`, no guard) and 166 (a versioned `site_specs`
> double-insert); it is silent on guarded files, on `ON CONFLICT` (167), on
> `WHERE NOT EXISTS` (184), on `UPDATE`-only files, on `doc_notes`-only files,
> and on the entire live pending queue.

Candidate (d) (auto-record on 23505) remains correctly rejected: a 23505 can also
be genuinely-wrong SQL colliding with unrelated data, and auto-recording would
mark broken SQL as applied — strictly worse than a replay. All three non-rejected
candidates (a, b, c) are now shipped.

### Verification

Exercised against a stub psql (`fake-psql.sh`, scratch ledger) so the destructive
paths could be driven without touching prod: sidecars listed-not-pending;
below-baseline files ignored; `--apply` stops on a forced 23505 and prints the
hint with the blocker **not** recorded; `--record-only` records, is a no-op on
repeat, and survives an apostrophe in the note; the run resumes to completion
once the blocker is recorded; `DOWN=1` refuses on both dry run and `--apply`;
flag validation. Against the **live** ledger: dry run now reports 3 pending
(was 5 — the two sidecars removed), 4.8s; the already-recorded no-op and the
sidecar refusal both behave, and `count(*) WHERE applied_by='record-only'`
stayed 0, confirming the probes wrote nothing.

> **[UNEXERCISED]** The `--record-only` **INSERT** has not run against the
> production ledger — deliberately. The only three pending files (177, 178, 179)
> are other threads' and are *genuinely* pending; recording them is exactly the
> harm this file warns about. First real out-of-band apply should use it and
> confirm the row.
>
> **CORRECTED 2026-07-22:** all three of 177/178/179 had been applied out of band
> (unrecorded) by the morning of 07-22 — their owner threads applied them between
> 07-21 and 07-22 without a ledger row. So the "genuinely pending" read was true
> when written and stale by the next morning; this is the trap's exact signature,
> and this thread recorded all three after verifying artifacts live. Do **not**
> read a past "genuinely pending" note as still-true — re-verify. Path now
> exercised (residual 2 CLOSED). Full story in **RECURRENCE 2026-07-22** below.

**Left alone:** ~~177/178/179 remain pending and unrecorded~~ (CORRECTED
2026-07-22 — now applied-and-recorded, see above). The two 175s and two 176s
remain as they are, per the numbering-collision note above. ~~The genuinely-pending
files as of 2026-07-22 11:48 BST are 186, 188, 194~~ (CORRECTED — all three were
applied out of band once v1.0.1149 shipped their checks that afternoon; see the
post-roll addendum. The queue is "Up to date" as of 2026-07-22 ~15:30 BST).

## RECURRENCE 2026-07-22 — it sprang SEVEN-fold, and residual 2 closed itself

The biggest instance yet, found by starting a session on this bug with the dry
run (exactly the discipline this file preaches). At ~10:00 UTC the runner
reported 5 pending; by the time each was verified the set had churned three more
times as concurrent threads committed new migrations. Verifying **every** pending
file against the live DB — the load-bearing step, never skipped — split them:

| file | reported | live check | verdict |
|---|---|---|---|
| 177 council tolerate_truncation | pending | 37/37 council `review_*` seats carry `tolerate_truncation=true` | **applied-unrecorded** → recorded |
| 178 news-listing JS PE | pending | `news-listing` js carries `hasServerRenderedItems` guard | **applied-unrecorded** → recorded |
| 179 news query-sourced items | pending | both news components v2 `fields`+`items`, templates render them | **applied-unrecorded** → recorded (RAISES on replay — a live halt removed) |
| 187 content_direction wireup | pending | `page-content-writer` prompt carries `current_page.content_direction`; column comment cites bug 025 | **applied-unrecorded** → recorded |
| 190 contact_form_undeliverable | pending | check present in `completeness-discovery-agent` run_checks | **applied-unrecorded** → recorded |
| 193 adoption_locked rename | pending | `build-site-planner` config carries both `site_has_no_current_plan` + `adoption_locked` alias | **applied-unrecorded** → recorded (by a concurrent session, in a race — see below) |
| 189 tool-fabrication gate | (cleared) | — | recorded by a concurrent session at 10:06 |
| 191 diagnose-agent resources | pending → cleared | resources now `250m/512Mi` (was `100m/256Mi`), ledger `applied_by='clients_user'` | **applied for real** by its owner (043) via the runner — healthy path |
| 186 truncated_component check | pending | check ABSENT from run_checks | **genuinely pending** — left (046, image-gated) |
| 188 backend_entry_orphaned check | pending | check ABSENT from run_checks | **genuinely pending** — left (017, image-gated) |
| 194 model_directory checks | pending | both checks ABSENT from run_checks | **genuinely pending** — left (Phase D, image-gated) |

**Seven applied-but-unrecorded files were live in the queue at once**, de-armed by
at least two–three sessions working concurrently within one hour. **Residual 2
closed itself in the process:** the `--record-only` INSERT is no longer
theoretical — it is in active fleet use, and a genuine race (a concurrent session
recorded 193 seconds before this thread tried) exercised the `ON CONFLICT DO
NOTHING` no-op exactly as designed.

**What this tells us, sharpened.** The process rule ("whoever applies, records")
has now failed at three events — 3 files (07-16), 3 files (07-20), **7 files
(07-22)** — across ~10 distinct workstreams. The trend is **up, not down**, four
days after the rule and the tooling both landed. This is no longer evidence that a
mechanism *might* help; it is proof the rule cannot hold in a repo where threads
routinely apply config-half migrations out of band (safe-before-image, live-on-
apply) and hand off mid-task. **The dry run must be run per session, not per day.**

### Post-roll addendum — the prediction held within hours (v1.0.1149, same day)

The three "genuinely pending" files above (186, 188, 194) were image-first seeds
waiting on a chassis image carrying their Go check. That image (**v1.0.1149**,
pod started 13:56 UTC) rolled a few hours later — pod-grep confirms all three
checks compiled in (`truncated_component` 9, `backend_entry_orphaned` 10,
`missing_model_directory_section`/`_page` 4/4). Within ~90 minutes **all three
seeds were applied out of band**, exactly the re-arm window the recurrence
predicted:

| file | how it landed | ledger | trap? |
|---|---|---|---|
| 194 | applied by owner | `manual-single-file` 14:25 | no — recorded |
| 186 | applied out of band | `record-only` 15:23 (concurrent session) | no — recorded |
| 188 | applied out of band by 017 | **none** → recorded by 007 thread 15:2x | **YES** — 188 RAISES on replay; a live halt, de-armed |

So one of three re-armed the trap on the way in (188). The queue then reached
**"Up to date — no pending migrations"** — a rare fully-clean state. The lesson
compounds: an image roll is not the end of the image-gated seeds' risk, it is the
*start* of their highest-risk window, because that is precisely when owners apply
the config half by hand. Re-run the dry run **after** any chassis roll, not just
at session start.

### A SAFE design for residual 1 (the halt), for the first time

The handoff called auto-record "unsafe" (candidate d) because a 23505 can be
genuinely-wrong SQL, and rightly rejected it. But the 2026-07-22 corpus shows the
failing-replay cases are not one class — they are three, and only ONE is
ambiguous:

- **Class A — self-declaring idempotent** (177, 187, 190, 193): a `WHERE … NOT
  LIKE` / `NOT (checks ? x)` guard makes a replay a 0-row no-op. Never halts.
  Harm is only that it sits "pending" for ever (and an *unconditional* rewrite
  like 178's `js_content` would clobber a later hand-edit on replay).
- **Class B — self-declaring guard-RAISE** (186, 188, and 179's precondition):
  `IF checks ? x THEN RAISE EXCEPTION 'NNN: x already enabled'`. Halts, but with a
  message the migration's OWN author wrote to mean "I have already run."
- **Class C — unguarded** (151, the 2026-07-16 culprit): a bare `INSERT` → raw
  Postgres 23505. This — and ONLY this — is the ambiguous case candidate (d) was
  rejected for: already-applied vs genuinely-wrong-SQL colliding with live data.

Two buildable, safe options fall out, both fleet-wide runner changes (shell-only,
live-on-commit) → **design/council-review first; owner = travelling-docs. Do not
build unilaterally.**

1. **Savepoint-probe in dry run (recommended — needs no new convention).** For
   each pending file, run its SQL inside `BEGIN … ROLLBACK`. If it RAISEs a
   Class-B "already"-guard, or a guarded file affects 0 rows, print **"LIKELY
   ALREADY APPLIED — verify & `--record-only`"**. This turns the *silent* Class-A
   files (the insidious ones that never halt) into loud dry-run warnings, keeps
   the human in the verify loop (candidate d stays rejected), and leverages guards
   that already exist. Risk to weigh: a migration with a non-transactional side
   effect (none seen in the corpus — all pure, transactional SQL, but audit
   before building).
2. **Auto-skip-and-record for Class B only, gated on a standardised sentinel.**
   Have already-applied guards RAISE a magic prefix (e.g. `-- ALREADY_APPLIED`),
   and let the runner record-and-continue on exactly that signal, never on a raw
   SQLSTATE. Safe, but needs the sentinel convention rolled across existing
   migrations first — heavier, and that rollout is itself the design work.

Until one lands, the halt stays a "die-loudly-with-a-fix-in-the-message" design,
which remains defensible — but the dry-run-per-session discipline above is now
mandatory, not advisory.

## References

- Travelling-docs `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` **T23** (discovery + backfill detail).
- Empty-sections `HANDOFF_2026-07-16_continue_here.md` **§7** (owner's record, updated with the resolution).
- `scripts/migration/run-migrations.sh` header comment (the guard-DO-block convention this trap violated).
