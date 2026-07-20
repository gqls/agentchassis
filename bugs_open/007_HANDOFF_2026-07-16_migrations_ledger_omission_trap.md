# HANDOFF — applied-but-unrecorded migrations block the runner (ledger-omission trap)

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

## References

- Travelling-docs `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` **T23** (discovery + backfill detail).
- Empty-sections `HANDOFF_2026-07-16_continue_here.md` **§7** (owner's record, updated with the resolution).
- `scripts/migration/run-migrations.sh` header comment (the guard-DO-block convention this trap violated).
