> ## ⚠ SUPERSEDED — the bug is CLOSED. Read `HANDOFF_2026-09-03_closed.md` in this directory.
>
> Nothing in this file's open list remains. What is left is three DECISIONS, laid out there.

> ## ⚠ SUPERSEDED 2026-09-02b — read `HANDOFF_2026-09-02b_continue_here.md` in this directory.
>
> Its open list is shorter than this one's: exactly ONE item remains (the CronJob firing on its
> schedule). Everything else here is either done or already stale.

> ## ⚠ SUPERSEDED 2026-09-02 — read `HANDOFF_2026-09-02_continue_here.md` in this directory instead.
>
> Four of the five open items below closed themselves over the following week, so this file's
> "do these next" list will send you chasing work that is already done: the audit TIMEOUT was
> transient (findings are being written again), the kustomize overlay is applied in-cluster, and
> the cursor has been running correctly on the SCHEDULED path — including a second site that
> cycled to completion. The evidence and the three genuinely remaining items are in the newer file.
> Everything below is accurate **as of 2026-08-26** and is kept for the trail.

# HANDOFF — 2026-08-26 (late evening) — `bugs_open/394`, render-audit coverage cursor

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/`
**Bug:** `bugs_open/394_HANDOFF_2026-08-25_webdesign_render_audit_tail_is_71_pages_and_growing_and_the_truncation_row_has_no_reader.md`
**Council:** `f67593f5-90cb-4a35-9cc0-926254645192` — **APPROVED** at round 3 (R1 and R2 both REVISE; both found real things).

---

## 1. State in one paragraph

The render audit used to take the **same first `max_pages` pages every run**, so anything past the
cap had never been audited and never would be. It now rotates on a keyset cursor, and the pages
carrying an open `contrast_failure` ride in **every** run so repair-grading keeps the 3-day cadence
migration 469 was commissioned to buy. **Code is LIVE** (verified at the binary on both replicas),
**migration 660 is APPLIED**, and the rotation is **PROVEN AT THE ARTEFACT** over two live runs.
One defect found by that live run is **fixed in code but NOT yet rolled**. The audit's *browser
round-trip* is currently timing out — that is a different subsystem and is the main open thread.

## 2. What is DONE and verified

| thing | evidence |
|---|---|
| Go code live on the fleet | capability probe, **both** replicas: marker `render_audit_page_cursor`=3, positive control=1, negative control=0 |
| migration 660 applied 22:20:40Z | `CREATE TABLE`, `UPDATE 1`, both `DO` verify blocks, `COMMIT` |
| pre-change backup exists | `agent_definitions_backup`, `snapshot_taken_at` 22:20:40Z, **`backup_has_flag = f`** (i.e. genuinely pre-change) |
| `design-critique-agent` NOT flipped | migration's own verify arm asserted it; re-checked = 0 |
| **run 1** hit all ten predictions | row 22:25:20Z — `cursor` / 151 / 60 / `index` → `tool-head-architect`, priority 3, dropped 0, not-live 0 |
| **run 2 is the proof** | row 22:32:36Z — `window_first = tool-html-minifier`, the page that had **never** been audited; `window_last = tool-entropy-meter-guide` at **`nav_order 200`**, the guide band unreachable at any cap below 98 |
| cursor advances | `(100, tool-head-architect)` → `(200, tool-entropy-meter-guide)` |
| commissioned reader built | `cmd/config-key-audit/rendertruncation.go`, 3 alarm arms + a not-live report line, **all mutation-proven**; registry flipped to `consumed` |

**Say it precisely:** the cursor **selects and sends** the right pages — that is what the durable
row proves and it is the half this change owns. It does **not** prove the browser measured them.

## 3. ⚠ OPEN — do these next, in this order

### (a) Roll an image — one committed fix is not live
`faf4872ce` fixes the cursor key: it was `params.ExecutionContext.Sender.AgentType`, the
**dispatcher's** identity, not the running agent's. Live consequence, observed: the hand-run wrote
its cursor under `agent_type='generic'` while the same run's truncation row said
`render-audit-agent`. Keyed that way, **one caller keeps a separate cursor per dispatch path**, so
the scheduled run and a hand run each restart from the top. Now keyed on
`runningStepProvenance(params)` — the same resolver `LogActionFindings` uses.

**Until that rolls**, expect the scheduled rotation to write its own cursor row under whatever
`Sender.AgentType` the scheduled topic carries. Not damage — split coverage progress.

**After the roll**, tidy the stray row (it is the only one, and it is honest history, so decide
rather than reflex-delete):
```sql
SELECT * FROM render_audit_page_cursor;          -- expect agent_type='generic' from 2026-08-26
-- optionally re-key it so the progress carries into the scheduled path:
UPDATE render_audit_page_cursor SET agent_type='render-audit-agent' WHERE agent_type='generic';
```

### (b) The audit step TIMES OUT — the biggest open question, and NOT this change
Both live runs ended `complete_error`, `{"message": "Request timed out (code: TIMEOUT)",
"failed_step": "audit"}`, ~3 minutes after dispatch. The render-audit adapter pod is **healthy** —
it completed a robot-hands.com audit at **22:00:54Z** and replied — but logged **nothing** from
either of my two dispatches. So the request is produced and the adapter never sees it.

Page selection is unaffected (both runs wrote their durable row and their cursor correctly). But
until this is resolved **no findings are written and nothing is graded**, so the coverage gain is
currently theoretical at the measurement level.

Start here:
```bash
kubectl -n ai-persona-system logs render-audit-adapter-<pod> --since=30m     # THE POD, not -l app=
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
  "SELECT current_step, status, collected_data->>'__step_error' FROM orchestration_states
    WHERE correlation_id='<corr>';"
```
⚠ `kubectl logs -l app=render-audit-adapter` returns **another service's lines** (one image, every
label). Read the pod.

Worth checking whether this predates today: another session recorded a **fleet-wide dispatch stall
~15:25Z** and a GitHub Actions outage on 2026-08-26.

### (c) Re-apply the optional-key-budget kustomize overlay
`rotate_coverage` took `request_render_audit` to **7** optional keys (budget N=10). The repo
literal is updated (`deployments/kustomize/services/optional-key-budget-check/base/check.py`), but
**the cluster keeps the old literal until the overlay is applied**. The parity test
(`go test ./cmd/config-key-audit/ -run BudgetCron`) is green in the repo and cannot see the cluster.

### (d) The reader has no CronJob yet
`cmd/config-key-audit --render-truncation` exists, is tested and is registered as the `consumed`
reader — but nothing runs it on a schedule. Clone the `ungraded-completions-check` kustomize
service. Until then the registry says `consumed` and no cron consumes it.

### (e) The acceptance arm, once (b) is fixed
151 pages / 57 rotation slots = **3 runs** for a full cycle. Union query is in RUNBOOK §1-adjacent;
grade the union of `audited_paths` over one cycle, and note that a page added mid-cycle may
lawfully wait one cycle.

## 4. Traps this lane paid for — read before touching these things

- **`snapshot_agent` has TWO overloads writing to DIFFERENT TABLES.** A bare literal is ambiguous
  and aborts the migration. The two-arg form writes `agent_definitions_backup` and **returns the
  SOURCE row's id**, so verifying in `agent_definitions` reads `0` and looks like a no-op. LANDMINE
  added.
- **NEVER parse a `contrast_failure` `item_key`.** A selector may contain `#`, and so may a page
  URL — `idea.uk` has BOTH `/tools.html` and `/tools.html#audience-check` ACTIVE with 35 open rows.
  Match forward with `workItemKey("contrast_failure", path+"#")` + `HasPrefix`, longest path first.
- **`sql_for_agents` numbers are a shared sequence with no reservation.** Mine was renumbered
  **twice** (646 → 649 → 660). Take `max+1` at the moment you APPLY, not when you write.
- **A test can only discriminate what its fixture varies.** `renderAuditParams` sets
  `Sender.AgentType` and the running agent to the same literal, which made twelve tests unable to
  see the key defect. Four `WRONG_CALLS` entries from this lane, all 2026-08-26.

## 5. Key facts, dated

`[MEASURED 2026-08-26]` webdesign.co.uk **151** live pages (146 earlier the same day, 131 on 08-24
— it is growing fast). Bands: `0..90` 6 nav pages · `100` 94 tools · `200` 48 `tool-*-guide` · `201` 1.
Two callers: `render-audit-agent` cap 60 (rotating), `design-critique-agent` cap 8 (deliberately
prefix, manual-only, acknowledged in `render_truncation_acks.json`). At cap 8, **25 sites** truncate.
Open `contrast_failure`: **116** fleet-wide, **3** on webdesign, **1** naming a no-longer-live page.

## 6. Also from this session

- **`bugs_open/359`** (retired pages still serving) — I validated it (**7 of 39** serving, controls
  held), then **yielded** the lane to a session that opened it two minutes after me; contributed
  two `CheckResult.Resolved` LANDMINES into their dir. They have since built the detector.
- **CONTRIB into `apis_uk_bees_homepage`** — their `SUBJECT_MISSING_ON_REPEATED_COMPONENT` registry
  entry puts its prose under `note` where the checker reads `why`, so
  `TestShippedRegistryIsSelfConsistent` is **red at HEAD**. Theirs to fix; not mine to edit.
- Register **VIZ-019** + index row.

## 7. Commit trail

`95a04168c` cursor · `72b16391b` R1 fix (forward match) · `ea08c831d` R3 advisories +
`Council-Reviewed` · `c71b46be0` 660 applied + landmine · `faf4872ce` **identity fix, not rolled** ·
`a3610ea23` artefact proof. Reader: `cmd/config-key-audit/rendertruncation.go`.
