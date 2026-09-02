> ## ⛔ SUPERSEDED 2026-09-02 — read
> `docs/agent_docs/docs024_key_docs_latest/deferred_work_item_park/HANDOFF_2026-09-02_continue_here.md` INSTEAD.
>
> Kept for the record; everything in it was true when written. What changed: the standing residual
> this file lists under "WHAT IS NOT DONE" — *nothing stops a raw UPDATE SET status='deferred'* —
> is now **closed in code** by migration `690` (register **WII-037**). ⚠ That migration is BUILT and
> TESTED but **NOT YET APPLIED**, so the hole is still open in production. The new handoff leads
> with the apply recipe.

# HANDOFF 2026-08-26b — `bugs_open/396`: everything is LIVE and RE-PROVEN on the new chassis `v1.0.1345`. Nothing is blocked. One contribution is out to another lane.

**Supersedes `HANDOFF_2026-08-26_continue_here.md`** (same directory), which is kept for the record.
Read this box; everything below it is evidence, recipe or background.

> ## STATE — all re-verified AFTER the 20:24Z roll, not carried forward
>
> | piece | state | how it was proven, today, post-roll |
> |---|---|---|
> | `honour_site_lock` arm in `LoadWorkItemsAction` | **LIVE** | binary probe, **both** replicas of `v1.0.1345` |
> | `sites.lock_except_item_ids` (migration `632`) | **APPLIED** | column present; symbol in both binaries |
> | migration `633` (the held config half) | **APPLIED + INTACT** | live selector carries the **exact parenthesised shape** |
> | `status_override` allow-list (council **APPROVED** `9c16eb83`) | **LIVE** | `WORK_ITEM_STATUS_OVERRIDE_REFUSED` in both binaries |
> | council round `175df761` r2 | **APPROVED** | `complete_approved` in `orchestration_states`, 09:00→09:10 |
> | the lock, honoured by the **real scheduler in production** | **OBSERVED** | two-arm live proof, §2 |
> | park verb `park_work_items()` (`621`, WII-034) | **applied, DEMOTED** | unchanged — see the old handoff §5 |
>
> **NOTHING IS BLOCKED AND NOTHING IS OWED ON THIS LANE.**
>
> **The one thing with a clock on it is not mine:** a CONTRIB is out to the `dispatch_throughput`
> lane about migration **`657`**, which rewrites the same selector query and is hand-applied
> **~12:00Z 2026-08-27**. Their text is correct; the note is about their guard. It is explicitly
> **not** a reason to delay their apply. §4.

---

## 1. What happened since the last handoff

Two findings this afternoon, then a chassis roll this evening which required re-proving everything.

**Both findings came from re-checking claims this lane had already written down** — not from new
work. That is the transferable part and it is why this section exists.

## 2. ✅ THE RESIDUAL IS CLOSED — the scheduler honours the lock in production

The previous handoff recorded an honest gap: the exception list had been exercised by running its
predicates verbatim inside a **rolled-back transaction** (6 → 0 → 1), but **nothing had watched the
real scheduler obey the lock.** It could not be forced, because `find_dispatchable_site` is
`ORDER BY wi.created_at ASC` **across the whole fleet** and ~1,400 items were queued ahead.

**It did not need forcing.** The **eight oldest dispatchable rows fleet-wide** sit on
`adversecreditmortgage.co.uk`, which is locked under
`locked_by = "portfolio_positioning: owner HALT 2026-08-18 …"`. A locked site heads the queue, so the
lock is exercised on **every tick**.

| arm | query | returns |
|---|---|---|
| **guard** | the live `find_dispatchable_site` text, **verbatim** | `agritec.uk` |
| **control** | the same text with **only the lock clause deleted** | **`adversecreditmortgage.co.uk`** |

**The two queries differ in exactly one clause, so that clause is what moved the answer.** Both arms
are read-only `SELECT`s; nothing was mutated. Re-confirmed after the roll (§5).

⚠ **The control is the half that matters.** Without it, "the guard did not return the locked site" is
equally consistent with the query being broken outright — a guard that never passes anything is
indistinguishable from a broken pipeline.

⚠ **THE TEST IS NOT ALWAYS AVAILABLE.** If the control returns an **unlocked** site, then no locked
site currently heads the ordering, the lock is not being exercised, and **neither arm means
anything.** That is *unavailable*, **not** *passed*. Do not record it as a pass.

**As of 20:35Z: 70 dispatchable items across 4 locked sites** are held by this clause (was 67/3 this
afternoon — another site was locked meanwhile). It is load-bearing right now, not latent.

## 3. ⚠ THE GUARD THIS LANE NOMINATED FOR MIGRATION AUTHORS WAS BLIND — and `633` blinded it

The approving council's one gating-level advisory was that `TestSiteLockExceptionSQLIsNotTheSelectorSpelling`
**cannot reach a migration author**, because migration SQL is text compiled against nothing. The
previous handoff answered that by nominating the `sites.locked_at` entry in `LANDMINES.md` as the
guard. **That entry's check could not discriminate.**

It was `... ->'config'->>'query' LIKE '%locked_at%'` → `HONOURS`/`IGNORES`: a **substring** test
against a clause whose **shape** holds the lock. Measured 2026-08-26 — all four return `HONOURS`:

| spelling | what it does | rows admitted |
|---|---|---|
| **A** the live clause | correct | 1,104 |
| **B** outer parens dropped | `AND` binds tighter than `OR` → status / attempt / retry / `depends_on` gates stop applying to every unlocked site | **15,683** — re-dispatches `complete`, `failed`, `cancelled` |
| **C** `OR COALESCE(...) IS NOT NULL` | never NULL → **lock off on every site** | releases the 70 held items, onto an owner HALT, next tick |
| **D** exception arm deleted | kills `lock_except_item_ids` silently | **no row count changes** — all locked sites have empty exception lists, so data cannot tell you |

**The check was not wrong when written.** On 2026-08-03 the clause was *absent*, and a substring test
detects absence perfectly. **Migration `633` — this lane's — made presence insufficient**, by making
the clause conditional. A check inherited straight across the change that invalidated it.

**Fixed in `LANDMINES.md`** (commit `455d86f53`; original left visible per convention): the
four-spelling table, a two-sided behavioural check, and an always-available `DO` block that executes
the **live query text** so it cannot drift from what runs. ⚠ Recorded honestly there — the block
catches **C** and **D**, and **cannot catch B**, whose damage lands on *unlocked* sites so the site it
returns is legitimately unlocked. **Nothing short of reading the parens catches B.**

Logged in `WRONG_CALLS.md`. The rule: **when you nominate an existing check as the guard for a new
failure mode, feed the new failure mode to the check and watch it fail** — the tell is that you
changed the thing the check inspects.

## 4. 📤 CONTRIB OUT — `dispatch_throughput` / migration `657`, applies ~12:00Z 2026-08-27

Having fixed the guard, I checked whether any **pending** migration touches that query. One does:
`657_selector_ranks_sites_by_loadable_work_HOLD.sql` (`bugs_open/413`), council-**APPROVED** r2, held
for hand-apply tomorrow.

- **Their query is CORRECT** — it carries `(s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))`
  properly wrapped, and their header names `config.query — NOT pre_query`. They read the landmine.
- **Their GUARD has the blind spot.** `657:201-209` tests each eligibility fragment with
  `position(v_frag in v_q)`. **Four of the seven fragments are OR-bearing and are listed WITHOUT the
  parens that wrap them** (the lock clause, `retry_after`, `approval_mode`, `depends_on`). Their own
  comment says each *"widens dispatch if dropped"* — **the precedence break widens dispatch without
  dropping anything.** It protects the *next* editor of that file, and against this it does not.
- **Their md5 precondition still HOLDS** — live selector md5 is `d6f98acdb5aec385d5eb4077eac530fc`,
  exactly what `657` expects, so it will apply cleanly.

Written into **their** directory, decision left to them, explicitly not a reason to delay:
`docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/CONTRIB_2026-08-26_from_deferred_work_item_park_657_guard_cannot_see_a_precedence_break.md`

## 5. THE POST-ROLL VERIFICATION — what a new chassis obliged, and the evidence

Chassis rolled **`v1.0.1341` → `v1.0.1345`**, pods up 20:24:56Z / 20:25:20Z. The lane's Go half lives
in that binary, so it was re-proven rather than assumed.

**Binary probe, BOTH replicas, present-control AND absent-control in the same run:**

| symbol | 5l8xd | 68t5h |
|---|---|---|
| `honour_site_lock` | PRESENT | PRESENT |
| `lock_except_item_ids` | PRESENT | PRESENT |
| `WORK_ITEM_STATUS_OVERRIDE_REFUSED` | PRESENT | PRESENT |
| `repairOutboundPageLinks` (**present-control**) | PRESENT | PRESENT |
| `zzzNotARealSymbol396zzz` (**absent-control**) | **absent** | **absent** |

⚠ **`build provenance` was NOT in range on either pod** — chassis produced **2.4 MB of logs in ten
minutes**, so the startup line had already scrolled past `--tail=20000`. **Empty there means "not in
range", never "unstamped".** The binary probe has no shelf life; use it.
⚠ The absent-control is the slow one (it must scan the whole binary) — **it timed out at 120 s on my
first attempt and I re-ran it alone at 240 s.** Do not accept a run whose control did not finish:
without it, every `PRESENT` is unfalsifiable.

**Config half, post-roll:** live selector names `lock_except_item_ids` ✓ **and carries the exact
parenthesised shape** ✓ (not just the substring — that is the whole lesson of §3).

**My new `DO`-block check, run for real against the live row:** `PASS: selector returned an UNLOCKED
site`. The check I wrote today works.

**One thing I could not attribute, stated rather than guessed:** the selector's `agent_definitions`
row shows `updated_at = 08-26 20:24:17`, ~40 s before the pods started. **I did not identify the
writer.** What is established: the `633` clause is present in its exact expected shape afterwards,
there is exactly **one** active non-snapshot row for this type, and the query text's md5 matches
`657`'s captured baseline. So the fix survived whatever it was. `[UNRESOLVED]`

**A theoretical revert path that did NOT fire, worth knowing about:**
`docs/agent_docs/sql_for_agents/052_build_pipeline_trigger.sql` — the seed for this agent — carries
the **OLD** query with bare `locked_at IS NULL` and **no** `lock_except_item_ids`. If anything ever
re-applies that seed, `633`'s config half is silently reverted. It is **not** in `schema_migrations`
and did not run at this roll. `[UNVERIFIED — I did not establish whether any path re-applies seeds]`

## 6. WHAT IS NOT DONE — stated so silence is not read as completion

- **The standing residual is unchanged: nothing stops a raw `UPDATE … SET status='deferred'`.**
  Short of a database trigger, nothing can. That is a platform change with its own review and it is
  the owner's call, not a side effect of a session. Recorded in `396` §6 and WII-036.
- **No parked row has been released.** 52 unstamped (mortgagecalculator 38, idea.uk 14), **62 stamped,
  60 carrying another lane's live `"un-park after rebuild verify"` condition.** ⚠ **Do not sweep them
  — ask the holders.** `unpark_work_items` is scoped to one `parked_by` for exactly this reason.
- **Spelling B cannot be caught by any string check**, including the one I added. Only reading the
  parens catches it. That is why any edit to this clause is a deliberate act, never a tidy-up.
- **I did not re-run the full `platform/orchestration/actions` suite** this evening. It was green
  (`ok … 5.426s`) after `a0ec90eb9` declared the finding code; other lanes have committed since.

## 7. ⚠ THE TRAPS — unchanged, and still the expensive ones

- **THE LOCK IS ENFORCED AT EXACTLY ONE AUTOMATED GATE.** `find_dispatchable_site` selects a **SITE**;
  `LoadWorkItemsAction` runs next and **never checked `sites.locked_at`** until `396` added the
  opt-in `honour_site_lock` arm. Applied config-first, `633` converts a full site hold into **no hold
  at all** — on precisely the sites somebody deliberately locked. **Binary first, config held.**
- **`load_work_item_actions.go:134` LOOKS like a second gate and is not.** It sits inside
  **`WriteBuildItemsAction`**, and **its log line misnames its own function**:
  `"LoadWorkItemsAction: site is locked, skipping"`. That string cost the previous session an hour.
- **The selector's SQL is under `config.query`, NOT `config.pre_query`.** A migration reading the
  wrong key patches nothing and reports success.
- **The two spellings of the lock rule are deliberately different and must stay different.** The Go
  fragment is per-site and `$1`-parameterised; the selector is a cross-site scan with no `$1`, spelled
  against the joined alias. `TestSiteLockExceptionSQLIsNotTheSelectorSpelling` fails if anyone merges
  them — a council reviewer already tried to, in their head.
- **`--record-only` REFUSES a `_HOLD` file** as an uppercase-suffixed sidecar. Held migrations are
  invisible to `schema_migrations` unless the row is hand-written, so **"was the held half applied?"
  can only be answered from live config.**
- **396 IS A DUPLICATE NUMBER.** The other `bugs_open/396_…_a_design_run_erases_every_appended_css_repair…`
  is a different bug in a different lane. **Resolve by slug; `git log` the FILE PATH, never the number.**

## 8. WHERE EVERYTHING LIVES

- **Bug:** `bugs_open/396_HANDOFF_2026-08-25_work_items_parked_at_deferred_with_a_named_handler_are_undispatchable_unrefilable_and_carry_no_provenance.md`
  — **§6d is today's state**, §6c the morning's, §6b the corrected direction.
- **This lane:** `docs/agent_docs/docs024_key_docs_latest/deferred_work_item_park/` — PLAN · NOTES
  (append-only, newest at the bottom — **the cold-start read**) · RUNBOOK · README_where_we_are
  (owner prose) · this handoff · `submission_396_*.json`.
- **Register:** **WII-036** (the exception list) · **WII-034** (the park verb, amended).
- **Code:** `platform/orchestration/actions/work_items_common.go` (`siteLockExceptionSQL`,
  `workItemStatusOverrideAllowed`), `load_work_item_actions.go` (`LoadWorkItemsAction`),
  `site_lock_exception_test.go` (4 tests), `status_override_allowlist_test.go`.
- **Migrations:** `621` · `632` · `633_..._HOLD` (each +ROLLBACK).
- **Councils:** `9c16eb83` APPROVED · `ed821065` REVISE (right, and acted on) · `175df761` r1 REVISE,
  r2 **APPROVED**.
- **Today's commits:** `455d86f53` (landmine correction + WRONG_CALLS + lane docs + `396` §6d).

## 9. IF YOU ARE PICKING THIS UP COLD — do this first

1. **Is the fleet up?** `GROUP BY success` is the whole query — an ungrouped `count(*)` reads healthy
   through an outage, because a failing endpoint produces *more* traffic:
   ```sql
   SELECT provider, success, count(*), to_char(max(created_at),'MM-DD HH24:MI') AS newest
   FROM llm_call_log WHERE provider='anthropic' AND created_at > now() - interval '30 minutes'
   GROUP BY 1,2 ORDER BY 3 DESC;
   ```
2. **Is the lock still honoured?** Run the `DO` block in the `sites.locked_at` entry of
   `LANDMINES.md` — it executes the **live** query text, so it cannot drift.
3. **Then stop.** There is no owed work on this lane. If you want the highest-value open thread in
   this area it is the standing residual in §6 — and that needs an owner decision, not a session.
