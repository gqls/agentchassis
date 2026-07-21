# HANDOFF — applied-but-unrecorded migrations block the runner (ledger-omission trap)

> ## STATUS 2026-07-21 — STILL OPEN, but the tooling shipped and is LIVE
>
> **Cold start:** read this banner, then jump to **"FIXED 2026-07-20 — the tooling
> landed"** near the bottom for what changed and how it was verified. The
> 2026-07-16 sections below are the original diagnosis and remain accurate history.
>
> **What is done and LIVE** (commit `a51333fd7`; shell only, so live on commit —
> no image roll): `run-migrations.sh` now has `--record-only <file> --note "<why>"`
> to register an out-of-band apply; a failure message that names the already-
> applied cause and prints the recovery command; sidecar (`_ROLLBACK`/`_VERIFY`)
> exclusion; a dry run cut from >120s to ~5s; and a refusal when the DB is
> unreachable. These are fix candidates **(a)** and **(b)** from below, plus three
> defects **(e)(f)(g)** found while taking the baseline dry run.
>
> **Why it is still OPEN, not moved to `/bugs_closed/`** (the bar there is
> *fixed AND live AND not reproducible*):
> 1. **The runner still HALTS on an already-applied-but-unrecorded migration.**
>    Fix (b) made that halt *informative* and gave a one-command recovery — it did
>    **not** make the runner auto-detect-and-skip (candidate (d), auto-record, was
>    deliberately rejected as unsafe). Recording is still a manual, opt-in step, so
>    a thread that forgets to record can still gate the queue for everyone. The
>    *diagnosis cost* (the 3-day misread) is fixed; the *block itself* is mitigated,
>    not eliminated — the symptom is still reproducible.
> 2. **The `--record-only` INSERT path is UNEXERCISED against the production
>    ledger** — deliberately: the only pending files were other threads' and
>    genuinely pending, so there was nothing safe to record. The next real
>    out-of-band apply should use it and confirm the row lands.
> 3. **Fix candidate (c)** — the unguarded-`INSERT` lint — is **not built**
>    (optional, preventive; left out to keep the change reviewable).
>
> **What IS eliminated:** near-miss (e) — the runner treating 180's `_ROLLBACK.sql`
> as a pending migration and reverting bug 024 — is closed and live.
>
> **Next actions for a resuming chat**, in priority order:
> - (nothing forces action — this is a mitigated landmine, not an outage.)
> - When you next apply a migration by hand: use `--record-only` and confirm the
>   ledger row, closing residual 2 above. That is the cheapest way to exercise it.
> - Optional robustness: build candidate (c) lint; or reconsider auto-detect
>   (skip-if-already-applied) now that the message path exists — note this changes
>   the runner from "die loudly" to "carry on", which the 2026-07-16 analysis
>   was wary of. Not obviously right; decide before building.
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

**NOT done: fix candidate (c)**, the unguarded-`INSERT` lint. Still wanted, still
optional; left out to keep this change reviewable. Candidate (d) (auto-record on
23505) remains correctly rejected.

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

**Left alone:** 177/178/179 remain pending and unrecorded — other threads' work,
and applying someone else's migration can violate an image-first ordering. The
two 175s and two 176s remain as they are, per the numbering-collision note above.

## References

- Travelling-docs `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` **T23** (discovery + backfill detail).
- Empty-sections `HANDOFF_2026-07-16_continue_here.md` **§7** (owner's record, updated with the resolution).
- `scripts/migration/run-migrations.sh` header comment (the guard-DO-block convention this trap violated).
