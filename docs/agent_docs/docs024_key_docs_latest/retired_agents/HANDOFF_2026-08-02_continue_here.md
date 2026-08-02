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

1. **Decide `report-builder`.** Left active deliberately. It is wired but its queue
   has been empty; that is a product question (are reports still a thing?), not a
   cleanup one.
2. **`intake-orchestrator` + `site-classifier`** — the last two genuinely orphaned
   agents of the old shape. Same evidence standard as §2. Use `retired_agents/` as
   the pattern: back up and commit *first*, then `is_active=false, deleted_at=now()`.
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
