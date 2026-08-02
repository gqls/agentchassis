# HANDOFF — 2026-08-02 late — build path, agent retirements, and what is left

**Start here for a new thread.** This covers a session that ran three phases:
finishing `bugs_closed/165`, answering whether the intake path is superseded, and
retiring four builder agents. Everything owed is **live and verified on
v1.0.1233**. Nothing is half-done.

Companion doc: `bugfix_165_reconciliation_deletes/HANDOFF_2026-08-02_continue_here.md`
(that lane is closed; read it only if you touch prune floors).

---

## 1. Live state — what is running right now

**Chassis `v1.0.1233`.** Everything this thread shipped is on it, pod-grepped on
both replicas with positive **and negative** controls:

| shipped | proof |
|---|---|
| completeness floors on all 4 reconciliation deletes | `bugs_closed/165`, 3 of 4 branch-pairs induced in production |
| `Reason()` required `aftermath` clause (corr `22cdef56`, APPROVED) | old **fused** literal reads **0**, new format string reads 1 |
| operator refusal text repointed to `bugs_closed/165` | `(bugs_open/165)` reads **0**, `(bugs_closed/165)` reads 1 |

**The builder menu is now two agents:**

| agent | state | why |
|---|---|---|
| `pageflow-builder` | **ACTIVE** | kept by owner instruction; `domain-research-classifier`'s prompt pins `recommended_builder` to it, so it must resolve |
| `report-builder` | **ACTIVE** | **saved during pre-flight** — see §3, this is the important one |
| `multipage-website-builder` | retired | 2 rows |
| `content-site-builder` | retired | |
| `landing-page-builder` | retired | |
| `website-builder` | retired | |

All retirements are `is_active=false, deleted_at=now()`. **Rows kept** — restore is
one `UPDATE` (`RESTORE_multipage-website-builder.sql` case 1). Full JSON backups in
this directory, both committed *before* the DB change.

## 2. The build path, measured — it was NOT superseded, it was re-plumbed

The live site-build flow, evidenced from `site_work_items.handler_agent` and
`site_specs.created_by` (**both durable, no reaper**):

```
domain-submitter          entry point. Spawns NOTHING — it files a WORK ITEM
  -> domain-research-classifier -> domain-strategist -> vertical-exemplar-researcher
  -> site-design-planner -> build-briefing-agent -> build-site-planner
```

**`build-site-planner` is the builder now** (`read_specs`, `plan_site`,
`validate_plan`, `reconcile_site_plan`, `write_site_plan`, `sync_pages`,
`populate_nav`, `emit_design`, `emit_imagery`). The architecture moved from
*orchestrator-spawns-a-builder* to *work-item routing*; it was not abandoned.

**Genuinely superseded and safe to retire if you want to finish the job:**
`intake-orchestrator` and `site-classifier` — both have 0 work items, 0 specs, no
`scheduled_tasks` row, and nothing spawns them. **Not retired**, because nobody
asked and they are harmless; but they are the last of the old shape.

## 3. THE THING TO CARRY FORWARD — how `report-builder` was nearly retired

The owner asked to retire every builder except `pageflow-builder`, on my report
that **"no work item has ever named a builder"**. That query was true and the
inference was wrong, and it nearly cost a live agent.

- `SELECT ... FROM site_work_items WHERE handler_agent LIKE '%-builder'` is empty.
  That is a statement about the **current queue**, not about **dispatch**.
- `report-dispatch` is an **enabled scheduled task on a 90-second tick** whose loop
  claims `pipeline='reports' AND item_type='report_request' AND
  status='awaiting_report'` and spawns *the handler named on the item* — its own
  step description reads "Spawn the handler named on the item (report-builder)".
- **8 `client_system.agent_instances` rows** reference it by `template_id`. A hard
  `DELETE` would have failed on that FK.

> **An empty queue plus an enabled dispatcher is not a dead agent — it is a live
> one with nothing to do**, and every "has it run" check reads zero in exactly that
> state.
>
> **Before retiring anything: an absence of WORK is not an absence of WIRING.**
> Check for (a) a live `scheduled_tasks` row, (b) FK referrers, (c) a dispatcher
> that resolves the handler *at runtime from data* — not only for rows naming the
> agent.

## 4. What is next — nothing is blocking, these are choices

1. ~~**Decide `report-builder`.**~~ **DECIDED 2026-08-02 (owner): KEEP IT ACTIVE.**
   No action taken or needed — it was never retired. Do not re-raise this on the
   evidence that its queue is empty; that is the §3 trap, and the answer is
   already recorded. It stays wired: `report-dispatch`, 90-second tick, 8
   `agent_instances` rows.
2. ~~**`intake-orchestrator` + `site-classifier`**~~ — **DONE, see §9.** Retired
   2026-08-02 late. The evidence standard in §2 turned out to be *insufficient* for
   these two, and the reason is worth reading before you retire anything else.
3. **`bugs_open/173`** — a refusal on one page aborts a whole multi-page build.
   Owned by the B/C lane; this thread contributed the census.
4. **Two consumers route on error rather than failing** (`page-build-handler`,
   `tool-recreation-handler`) so a refusal is recorded while the pipeline reports
   success. Content is protected by construction; it is a *visibility* gap.
5. **`MEMORY.md` is ~23KB** against a hook that wants 17.1KB. I cut what was
   legitimately cuttable (closed-bug entries). Reaching the target means cutting
   the practice families, which the file's own owner ruling forbids ("~20KB is
   ACCEPTED; do not compact by moving the practices out"). **That conflict needs an
   owner decision, not another compaction pass.**

## 5. Landmines this session produced (all in `LANDMINES.md`, synced to `doc_notes`)

- **`orchestration_states` keeps terminal rows ~24 HOURS**, and whole-table
  `min(created_at)` says *20 days* because `CANCELLED`/`RUNNING`/`INITIALIZED` are
  not reaped. Bound **per status**, then re-source durable claims from a table with
  no reaper (`site_specs` goes back to 2026-02-25; `site_work_items` likewise).
- **`git diff | grep '^-[^-]'` cannot see a deleted markdown bullet** (`-- **x**`
  is two hyphens) — gate on `git diff --numstat`. Two siblings on the same entry:
  grepping a diff for a symbol counts **context** lines, and `git log -S` is
  occurrence-COUNT based so it misses an edit preserving the count — use `-G`.

## 6. Mistakes worth inheriting (the tally is the point)

1. **I cited a bug number without resolving it.** "Blocked on `bugs_open/092`" went
   into three durable documents; 092 had closed the same day the pointer was
   written. `who-owns.py <n>` costs 0.3s. **Citing a bug number is acting on it.**
2. **A deferral names a destination and nobody checks it was accepted.** 092 said
   the link-registry question was 165's; 165 said it was 092's. Neither owned it
   and both closed. **Write the item into the OTHER case's file.**
3. **I filed the retention landmine and then reasoned from the same table hours
   later**, concluding the intake path was superseded. It was not. Knowing the
   class does not protect you; changing the source does.
4. **A safety check written in a hurry answers a neighbouring question** — three
   spellings of that in one session, all failing *open*.
5. **My own council-submission text scored as usage** of the agents it argued
   about: 15 `orchestration_states` matches, all owned by `council-gate`. Check
   `owner_agent_type` before treating a `collected_data` match as evidence.

## 7. Verification commands (copy-paste)

```bash
# what is actually in the running binary — always with a NEGATIVE control
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "<string your change ADDED>"; \
   strings /app/agent-chassis | grep -c "<string it REMOVED>"'   # expect N, then 0
```

```sql
-- the builder menu as intake actually queries it (active_only defaults true)
SELECT type FROM agent_definitions WHERE is_active = true AND type LIKE '%-builder' ORDER BY type;

-- durable "who really runs" — NOT orchestration_states
SELECT handler_agent, count(*), max(created_at)::date FROM site_work_items
WHERE handler_agent <> '' GROUP BY 1 ORDER BY 3 DESC;
SELECT created_by, count(*), max(created_at)::date FROM site_specs GROUP BY 1 ORDER BY 3 DESC;

-- before retiring ANY agent: wiring, not work
SELECT name, enabled, interval_seconds, last_triggered_at FROM scheduled_tasks
WHERE target_agent_type = '<agent>' OR input_data::text LIKE '%<agent>%';
SELECT count(*) FROM client_system.agent_instances i JOIN agent_definitions d ON d.id=i.template_id
WHERE d.type = '<agent>';
```

## 8. Cold-start pointers

- Retirements + backups + restore: **this directory**
- Prune-floor lane (closed): `bugfix_165_reconciliation_deletes/`
- Cases: `bugs_closed/165`, `bugs_closed/135`, `bugs_closed/092`; open: `bugs_open/173`
- Register: **CTXA-025** in `docs026_concept_register/register/context-assembly.md`
- Councils, all APPROVED: `a54172b6` (site A), `c69e935a` (B+C), `22cdef56` (aftermath)

---

# 9. UPDATE — 2026-08-02, later: §4.2 is done, and it needed a fourth axis

**Retired:** `intake-orchestrator`, `site-classifier`. `UPDATE 2`, soft
(`is_active=false, deleted_at=now()`, rows kept). Backup and restore committed
**before** the DB change, as `5fe6b173a`. Full account in
`README_multipage-website-builder.md` § "Third retirement";
restore is `RESTORE_intake_path_orphans.sql` — **they are a pair, restore both or
neither.**

Post-checks, all clean: gone from an `active_only` lookup · builder menu untouched
(`pageflow-builder`, `report-builder`) · all 7 live-path agents still active · no
active config names a retired agent.

## The fourth axis — and §3's lesson has now bitten three times

§3 says *an absence of WORK is not an absence of WIRING*. These two needed one
more step, because **`intake-orchestrator` is an ENTRY POINT**: it is spawned by
an operator publishing to `system.agent.generic.requests`, so it has no referrer
*by definition* and reads as an orphan on every DB axis **whether it is dead or in
daily use**. Zero rows meant nothing here.

What made it safe was a **file-date comparison in `scripts/`**, not a query:
`090_new_build/…_intake_orchestrator.sh` last touched **2026-06-21**;
`020_build_pipeline/082_submit_domain_unified.sh` (**2026-07-30**) routes to
`domain-submitter` instead. The operator habit **moved**. Filed as a landmine
(`832c10330`): *an absence of WIRING is not an absence of a CALLER, when the
caller is a person with a script.*

## Two defects found in the landmine delivery path itself

Found while filing the above — the routine `landmines-sync.py --apply` step.

1. **A `##` heading silently cost TWO entries their delivery.** `HEADING_RE`
   matched `###` only, so a `##` heading was not a heading: its lines were
   absorbed into the *preceding* entry and its `**footprint:**` line overwrote
   that entry's own. Measured: **`UpsertPageForRole` had 0 `doc_notes` rows** —
   the pages-upsert landmine the 175 lane filed that morning could not reach
   anyone, because the 172 lane's `##` append hours later swallowed it. Two
   malformed headings existed; both fixed (`b88cb70f5`), parser hardened and
   proven by induction (`8781bd811`). 790 → 799 rows.
2. **`parse()` takes an `on_warn` and both callers passed none**, so every warning
   it has ever raised went to nobody. Now wired up and partitioned.

### THE FOLLOW-UP THIS LEAVES YOU — 12 landmines that reach nobody

Wiring the warnings up immediately surfaced what the silence was hiding: **12
entries are SKIPPED for having no `footprint:` line.** They are in `LANDMINES.md`
and in **no** `doc_notes` row, so no SessionStart hook, no council seat and no
agent has ever been shown one. See them with `./scripts/landmines-sync.py`:

```
A migration's verify block made of `SELECT`s cannot stop the `COMMIT`
Deleting a workflow step: the SUCCESS edge is the one you remember
Cloudflare answers `Python-urllib` with 403
`site_components` having rows does NOT mean the chrome works
A verbatim page is defined by its ROW COUNT
A splitting/extraction rule that reads only INLINE scripts is blind
Two migrations can carry the SAME number
A component template's `{{else}}` can INVENT a business fact
Decomposing an adopted site with NO `site_components` head
A structural parity test cannot see an ENCODING
A `page_rerender` regenerates SECTIONS only when `spec.reason` is set
Any regex you write against HTML will also match inside `<script>`/`<style>`
```

**Not fixed here, deliberately:** each needs a `footprint:` written by someone who
knows what it guards, and guessing would key a landmine to the wrong grep — which
is worse than the current honest absence. Two of them are load-bearing elsewhere
(the migration-verify one and the regex-inside-`<script>` one are both cited in
`MEMORY.md`), so this is not a backlog of stale entries. **A one-off fix is not a
class fix**: the parser now refuses to be silent, but these 12 predate it.
