# HANDOFF 2026-09-02 — `bugs_open/396`: the standing residual is CLOSED, LIVE AND PROVEN. **Nothing is owed.**

**Supersedes `HANDOFF_2026-08-26b_continue_here.md`** (same directory), kept for the record.
Read this box. Everything below it is evidence or recipe.

> ## STATE
>
> | piece | state |
> |---|---|
> | `status_override` allow-list — council APPROVED `9c16eb83` | **LIVE** |
> | `sites.lock_except_item_ids` — migrations `632` + `633` | **LIVE, and PROVEN in production** |
> | `honour_site_lock` arm in `LoadWorkItemsAction` | **LIVE** on `v1.0.1345`, both replicas |
> | park verb `park_work_items()` — `621`, WII-034 | applied, **DEMOTED** (see the old handoff §5) |
> | **`refuse_untraceable_park()` — migration `690`, WII-037** | ✅ **LIVE AND PROVEN 2026-09-02 16:16Z** |
>
> ## ✅ APPLIED 2026-09-02 16:16Z — NOTHING IS OWED ON THIS LANE
>
> The owner ran the apply by hand (the building session's harness classifier gates live schema
> changes, which is why it could not). `trg_site_work_items_park_provenance` is **attached and
> enabled** (`tgenabled='O'`); the ledger row is recorded `applied_by='record-only'`; **zero litter
> rows** remain from the self-test. The migration's post-check passed **6 assertions before
> COMMIT**, and the independent `_VERIFY` then passed **all 6 against the live trigger, exit 0**.
>
> ⚠ **The one thing still open is not work: council `dcd2b3c9`'s verdict has NOT been read.**
> Read it before writing `Council-Reviewed:` anywhere — `098` buckets an unread claim as MISMATCH,
> which is the coverage report's dishonesty surface.
>
> **The commands below are kept as the record of how it was applied, and as the re-verify recipe.**
>
> ```bash
> # 1. apply — the file SELF-TESTS and aborts before COMMIT if the guard does not fire
> kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
>   -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/690_refuse_untraceable_park.sql
>
> # 2. record it (the ledger row is a separate, human act)
> ./scripts/migration/run-migrations.sh --record-only 690_refuse_untraceable_park.sql \
>   --note 'applied by hand; post-check passed 6 assertions'
>
> # 3. prove it behaviourally — this file ends in ROLLBACK and touches no real row
> kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
>   -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/690_refuse_untraceable_park_VERIFY.sql
> ```
>
> ⚠⚠ **DO NOT RUN `run-migrations.sh --apply`.** `[MEASURED 2026-09-02]` **271 files are pending**
> and it takes **EVERY** one — that would push ~20 other lanes' migrations live at once. Apply this
> one file by hand, then `--record-only`. (`668` is also duplicated on disk today — two different
> files, same number, the documented trap, live.)
>
> **Withdrawal is one statement, immediate, no roll:**
> `DROP TRIGGER trg_site_work_items_park_provenance ON site_work_items;`
> The refusal message carries that line as its `HINT`, so anyone blocked by it at an awkward
> moment is told how to withdraw it without finding this file.
>
> **Council SUBMITTED** `dcd2b3c9-cf38-4887-803a-9df6e27dcefe` — verdict not read. Read it before
> claiming approval; `Council-Reviewed:` on an unread verdict is `098`'s dishonesty surface.

---

## 1. What `690` does, and the scope that keeps it safe

It refuses a **transition into** `status='deferred'` on a row with a **NAMED** `handler_agent`
unless the same write carries **both** `parked_by` and `parked_reason`, in **`spec` OR `result`**.

**The scope was set by a census, not by the bug file.** `[MEASURED 2026-09-02]`

| shape | rows | with provenance |
|---|---|---|
| `deferred` + **EMPTY** handler — the `bugs_closed/077` "shelf", **legitimate** | **2,656** | **0**, correctly |
| `deferred` + **NAMED** handler — this bug's shape | **257** | 87 (all `migration_389`), **170 without** |

The shelf class is a **different mechanism** with five live producers (`write_audit_findings_action`
in `filing_mode='record'`, `capability_gap`, `discovery_checks/remit.go`, `check_palette_contrast`,
`check_content_duplication`, `check_missing_tools`, `cmd/verifier-remit-check`). **A guard requiring
provenance on every `deferred` write would have refused all 2,656 and broken all five.** That was
this session's starting design and the census is what killed it.

## 2. Three things that only reading the source gave — each would have been a real defect

- **Provenance must be accepted in EITHER `spec` OR `result`.** Migration `389` stamps `spec`;
  `park_work_items()` stamps `result`, deliberately (WII-034: `refreshOpenWorkItemSQL` replaces
  `spec` wholesale, so a `spec` stamp is destroyed by the re-detection a parked row is most exposed
  to). **A guard reading only `spec` would refuse the sanctioned verb** — and reading only `spec` is
  this lane's own §8.1 misstep, the one that reported 62 fully-stamped rows as "no trace of any kind".
- **It must fire only on the TRANSITION in.** Gated on `OLD.status IS DISTINCT FROM 'deferred'`, so
  the **170** legacy unstamped rows stay writable and `work_item_retraction.go` can still drain
  them. Firing on every write to a deferred row would strand exactly the rows this bug is about.
- **The two Go files naming `deferred` beside a handler are READERS, not writers.**
  `work_item_failure_ladder.go` has it in a guard list of statuses a failure/completion write must
  NOT overwrite; `work_item_retraction.go` reads it to count parks being drained. Every actual Go
  writer of `deferred` pairs it with an empty handler, and `write_build_items_routing_test.go`
  **asserts** that. So no live Go path can trip this.

## 3. How it was tested — and why the mutation half is the part that counts

Three clean dry runs (the whole file with `COMMIT` swapped for `ROLLBACK`): **6 assertions, exit 0,
nothing installed.** Then two deliberate breakages, each caught by a **different** assertion:

| mutation | caught by | result |
|---|---|---|
| guard made inert (always `RETURN NEW`) | assertion 1 | *"an untraceable park was ACCEPTED"*, exit 3 |
| shelf exemption removed (guard too broad) | assertion 4 | *"the SHELF class was REFUSED"*, exit 3 |

**The second is the one that matters.** A one-sided "did it refuse?" test passes a guard that
refuses *everything* — which would break 2,656 live rows and five producers. This is the same
two-sidedness the lane learned on 08-26, when the guard it had nominated turned out to be a
substring test returning `HONOURS` on four different spellings.

The migration's post-check **induces** the refusal and aborts before `COMMIT` if it does not fire,
because **a verify block of bare `SELECT`s cannot stop a `COMMIT`** (`ON_ERROR_STOP` ignores a
non-empty result set). Synthetic rows only, deleted before commit, with a litter check.

⚠ **The dry run found a real defect in the TEST, not the guard.** CHECK constraint
`swi_no_handlerless_promotable` forbids an empty `handler_agent` in `triaged`/`approved`/`claimed`,
so a shelf row is **born** `deferred` and can never be updated into it. The first draft staged one
at `triaged` — a shape production cannot produce — and was rejected. Corrected.

## 4. What is NOT closed — stated so silence is not read as completion

- ~~**`690` IS NOT APPLIED.**~~ **APPLIED 2026-09-02 16:16Z and proven at the artefact.** What remains is only to READ council `dcd2b3c9`'s verdict — do not write `Council-Reviewed:` until you have.
- **It enforces presence, not truth.** A false `parked_by` still passes. Nothing short of review
  catches that, and it is a much smaller problem than an anonymous park.
- **It cannot attribute the 170 existing rows.** That information was never written.
- **No parked row has been released.** 170 unattributable; of the attributable ones, **60 carry
  another lane's live `"un-park after rebuild verify"` condition.** ⚠ **Do not sweep them — ask the
  holders.** `unpark_work_items` is scoped to one `parked_by` for exactly this reason.
- **Council verdict on `dcd2b3c9` not read.**
- **The `657` CONTRIB** (see the 08-26b handoff §4) was left with the `dispatch_throughput` lane;
  no reply checked this session.

## 5. ⚠ TRAPS — the expensive ones, unchanged plus two new

- **`run-migrations.sh --apply` takes EVERY pending file (271 today).** Scope by hand.
- **A migration number is not reserved by creating the file.** `691` was taken by another session
  while this work was in progress, and `668` is duplicated on disk right now.
- **`deferred` means two different things.** Empty handler = the legitimate shelf; named handler =
  this bug. Any query, guard or census that does not split on `handler_agent` is answering a
  different question from the one it looks like it is answering.
- **The lock is enforced at exactly one automated gate** (`find_dispatchable_site`'s `config.query`,
  NOT `pre_query`); `LoadWorkItemsAction` honours it only behind opt-in `honour_site_lock`.
- **`load_work_item_actions.go:134` LOOKS like a second gate and is not** — it sits inside
  `WriteBuildItemsAction` and **its log line misnames its own function**.
- **396 IS A DUPLICATE NUMBER.** The other `396_..._a_design_run_erases_every_appended_css_repair...`
  is a different bug in a different lane. **Resolve by slug; `git log` the FILE PATH.**

## 6. WHERE EVERYTHING LIVES

- **Bug:** `bugs_open/396_HANDOFF_2026-08-25_work_items_parked_at_deferred_with_a_named_handler_are_undispatchable_unrefilable_and_carry_no_provenance.md`
  — **§6f is today**, §6e the post-roll re-proof, §6d the production proof of the lock, §6b the
  corrected direction.
- **This lane:** `docs/agent_docs/docs024_key_docs_latest/deferred_work_item_park/` — NOTES
  (append-only, newest at the bottom — **the cold-start read**) · README_where_we_are (owner prose)
  · PLAN · RUNBOOK · this handoff · `submission_396_*.json`.
- **Migrations:** `690_refuse_untraceable_park.sql` (+`_ROLLBACK`, +`_VERIFY`) · `632` · `633_HOLD`
  · `621` (the park verb).
- **Register:** **WII-037** (this trigger) · **WII-036** (the site lock) · **WII-034** (the park
  verb). All three were corrected today — WII-034 and WII-036 both still stated this residual as
  open, and WII-036's status still said its config half was held.
- **Councils:** `9c16eb83` APPROVED · `ed821065` REVISE · `175df761` r2 APPROVED ·
  **`dcd2b3c9` SUBMITTED, unread.**
- **Commits:** `a027bf03b` (migration + sidecars) and today's docs/register commit.

## 7. IF YOU ARE PICKING THIS UP COLD

1. **Check whether `690` was applied** — it is the only open item:
   ```sql
   SELECT count(*) FROM schema_migrations WHERE filename LIKE '690%';
   SELECT tgname FROM pg_trigger WHERE tgrelid='site_work_items'::regclass
     AND tgname='trg_site_work_items_park_provenance' AND NOT tgisinternal;
   ```
2. **If applied**, run the `_VERIFY` (it ends in `ROLLBACK`) and read the council verdict.
3. **If not**, apply it with the recipe in the box at the top. Do not use `--apply`.
