# NOTES — bugs_open/116, link-integrity check coverage

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-03 ~21:00–21:30 — taking the bug, and re-measuring it before believing it

Session "bugfix 100" (renamed), working the owner's standing instruction: take the
next `bugs_open/` bug no other thread holds, verify it is still valid, fix it at
the framework level.

### Why 116 and not another

Cross-referenced all 53 open bug files against every session transcript modified
in the last 4 hours (`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`,
`find -mmin -240`, grepping each for `bugs_open/NNN` / `bug NNN` / `bugfix NNN`).
Eight open bugs had **no** live session mentioning them: 107, 113, 114, 116, 121,
132, 147, 170.

Discarded, with the reason:

| bug | why not taken |
|---|---|
| 132 | **Blocked, not unowned.** The fix is a Cloudflare edge change; the deployed worker's source is in no repo (`scripts/cloudflare/worker.js` exists but its miss-branch returns plain `Not found`, while the live edge returns the JSON blob — so the repo copy is not what runs). No `wrangler.toml`, no deploy path from this tree. |
| 113 | Renderer fix already shipped (`palette_specialised_slots.go`, 3096a55a6). Residual is per-site stylesheet re-renders owned by other workstreams — the file says "do not repaint them from here". |
| 114 | Handler now exists (`page-build-handler`) and 23 items have completed since 07-28; the residual overlaps `bugs_open/187`'s active lane. |
| 170 | Already fixed and committed (`e44e6dd06`) by its own lane; needs a roll, not a fix. |
| 107, 121, 147 | 107 is design-scope and fuzzy; 121's second half is inert-until-roll; 147 is a single-site content defect. |

116 is unowned, fleet-wide, and its fix is inherently a framework seam — which is
what the instruction asks for.

### The filing's headline claim is now FALSE, and it matters

`116` is titled *"the link-integrity audit has never run, on any site"*. **That is
no longer true.** Measured live 2026-08-03 ~21:15Z:

```sql
SELECT item_type, source, created_by, count(*), count(DISTINCT site_id) AS sites,
       min(created_at)::date AS oldest, max(created_at) AS newest
FROM site_work_items
WHERE item_type IN ('phantom_internal_link','dead_control','cta_names_unknown_destination')
GROUP BY 1,2,3 ORDER BY 4 DESC;
```
```
 cta_names_unknown_destination | discovery | generic                      | 61 | 6 | 2026-07-14 | 2026-08-03 10:37:47
 cta_names_unknown_destination | discovery | completeness-discovery-agent | 60 | 6 | 2026-07-14 | 2026-08-03 21:04:02
 phantom_internal_link         | discovery | completeness-discovery-agent | 54 | 5 | 2026-07-22 | 2026-08-03 21:04:02
 phantom_internal_link         | discovery | generic                      | 45 | 4 | 2026-07-18 | 2026-07-28 15:15:28
 dead_control                  | discovery | completeness-discovery-agent |  4 | 3 | 2026-07-17 | 2026-08-03 21:04:07
 dead_control                  | discovery | generic                      |  4 | 1 | 2026-07-18 | 2026-07-18 15:51:47
```

The checks ran **today**, on `gaswholesalers.com`, `vonc.com` and
`gamesdesign.co.uk`, and filed real findings. So the file's central factual claim
has expired and must be corrected in place before anything is built on it.

**[MEASURED] Note the DB item_types are singular** (`phantom_internal_link`,
`dead_control`) and one is renamed (`cta_names_unknown_destination`), while the
check names in config are plural (`phantom_internal_links`, `misdirected_cta`). A
grep for the check name against `site_work_items.item_type` returns zero and looks
like "never ran". That is a trap, and it is probably part of why the original
filing read the way it did.

### What actually runs the checks

`completeness-discovery-agent`'s `run_checks` step (`action: run_discovery_checks`)
carries 31 checks including all three link ones, scoped `site_id: site_record.site_id`
— a whole-site sweep, not per page.

It is reached only as a **spawn inside `improvement-loop`'s audit chain**:

```
load_audit_state → check_audit_due (conditional: audit_state.audit_due == true)
   → spawn_quality_discovery → call_quality_discovery
   → spawn_design_discovery  → call_design_discovery
   → spawn_completeness_discovery → call_completeness_discovery   <- the three link checks
   → spawn_design_audit → call_design_audit → spawn_site_review → call_site_review
   → record_audit_pass → triage_findings → …
```

So the audit is **whole-site, gated on `audit_state.audit_due`, and sits behind
four other spawned agents**. That shape matters for the fix: it is not a cheap
thing to hang off every page write.

### MISSTEP — I measured coverage the wrong way first, and the wrong number looked fine

My first coverage query counted **sites with link findings**:

```
 sites_total | ever_link_audited | never_link_audited
          37 |                10 |                 27
```

**That number cannot answer the question and I should not have run it.** A site
with no findings is either clean or unexamined, and this bug file's own §"Why it
is dangerous" says exactly that: *"The checks passing and the checks not running
are indistinguishable from outside."* I encoded the flattering reading into the
filter and it produced a confident, meaningless 10. Caught it before using it.

The disconfirming measurement exists and is durable — `improvement-loop`'s
`record_audit_pass` step writes it:

```sql
-- record_audit_pass, verbatim from the live agent definition
UPDATE sites SET settings = jsonb_set(jsonb_set(COALESCE(settings,'{}'::jsonb),
  '{maintenance_profile}', COALESCE(settings->'maintenance_profile','{}'::jsonb), true),
  '{maintenance_profile,last_audit}',
  jsonb_build_object('fingerprint',$2::text,'at',now(),'passes_at_fingerprint', …), true)
WHERE id = $1
```

Asked properly:

```sql
SELECT domain, status, (settings#>>'{maintenance_profile,last_audit,at}')::timestamptz AS last_audit
FROM sites ORDER BY last_audit DESC NULLS LAST;
```
```
 gaswholesalers.com | deployed | 2026-08-03 21:07:34
 vonc.com           | deployed | 2026-08-03 21:07:29
 gamesdesign.co.uk  | deployed | 2026-08-03 21:07:21
 finetuning.uk      | deployed | 2026-08-03 10:19:41
 … every other row NULL …
```

**4 of 37 site rows carry an audit stamp — and all four were stamped today.**
Excluding 17 `pool-*.internal` rows and `system.internal`, the real fleet is 19
sites, so it is **4 of 19**.

> **[UNVERIFIED] — do not state this as "15 sites have never been audited".** The
> `last_audit` key is written by a step whose description cites `171`, so the field
> may well be younger than the fleet. A NULL therefore means "not audited since
> this field existed", which is a weaker claim than "never audited". The
> introduction date needs establishing before any count is published. Recorded
> here rather than left as an assumption.

### The driver is off, and the row that would tell you so lies

```
 name              | enabled | interval | target_agent_type | last_triggered_at   | last_completed_at
 improvement-sweep | f       | 180      | improvement-loop  | 2026-05-02 10:11:07 | 2026-08-03 21:18:56
```

`enabled = false`, and `cmd/scheduler/main.go:360` `loadDueTasks` selects
`WHERE enabled = true` — so **the scheduler does not fire it.** Yet
`last_completed_at` advances every few minutes.

The explanation is in `improvement-loop`'s own `notify_scheduler` step:

```json
{"action":"query_database",
 "config":{"query":"UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'improvement-sweep'"}}
```

Every improvement-loop run stamps that row **by name**, whether or not the
scheduler dispatched it. So a fresh `last_completed_at` on a disabled task is not
evidence of anything. → LANDMINE candidate; see LANDMINES.md.

`improvement-sweep` is referenced by only three active definitions
(`improvement-loop`, `council-gate`, `fix-proposer`), and no enabled scheduled
task targets `improvement-loop`. **So nothing drives the audit automatically.**
Today's four audits were dispatched by hand.

### Where that leaves the bug

The filing's *headline* is stale; its *substance* stands, and its own candidate 1
is still unimplemented:

> owner, 2026-07-27, quoted in the file: *"whilst the improvement loop will return
> the checkers should run after every build or change I think"*

Diagnosis loop fired 2026-08-03 ~21:30Z to test the mechanism claim independently
before I assert it (`090` trigger; symptom states the mechanism and points at
`scheduled_tasks`, `cmd/scheduler/main.go:360`, the two workflow steps, and
`sites.settings->maintenance_profile->last_audit`, asserting no counts).
Intake corr `aadb9c93-62af-4676-993f-b741310c2371`; run corr
`54bf4506-5192-4528-8395-eb2c636a7fad`.

## 2026-08-03 ~22:10 — MISSTEP: I nearly filed a "latent double-dispatch" bug that does not exist

Having found that `improvement-loop`'s `notify_scheduler` step stamps
`last_completed_at` on the **disabled** `improvement-sweep` row by name, I built a
theory: `cmd/scheduler/main.go`'s in-flight guard reads
`last_completed_at >= last_triggered_at`, so a manual improvement-loop run that
stamps completion while a scheduler-dispatched run is genuinely in flight would
clear the guard and let the sweep fire again — a double-dispatch armed the moment
G1 flips the sweep on. It was a tidy, plausible, *filable* finding.

**It is wrong, and reading one function killed it.** `cmd/scheduler/main.go:287`
calls `stampCompleted` immediately after `fireTrigger`, and `stampCompleted`
(`:343-348`) advances **both** columns:

```go
`UPDATE scheduled_tasks SET last_triggered_at = NOW(), last_completed_at = NOW(), updated_at = NOW() WHERE id = $1`
```

Its own doc comment says why (`:337-342`): advancing both *"keeps the task out of
countInFlight and satisfies loadDueTasks' in-flight guard"*. So for a
message-firing task the guard is already satisfied at fire time by design — the
agent's redundant stamp cannot clear a guard that is never held.

**What survives is a diagnostics defect only:** a disabled scheduled task shows a
`last_completed_at` from a minute ago, because any improvement-loop run stamps it
by name regardless of what dispatched it. That misleads a reader about whether the
scheduler is firing; it does not misdirect the scheduler. Recorded as a LANDMINE,
**not** filed as a bug.

The cheap check that would have saved the theory is the one I eventually ran:
**read the function that writes the column before reasoning about who reads it.**

## 2026-08-03 ~22:20 — the conclusion: this is not a coding task, and building the fix would violate policy

Research over the docs corpus (two passes) established that all four of this bug's
fix candidates are closed off. Full table with citations is now in the bug file's
STATUS block; the load-bearing ones:

- **Candidate 1 (per-build detection) is forbidden by the platform's own written
  policy**, not merely risky. `TriageDetectedItemsAction` — the only thing that
  promotes `detected` → `triaged` — lives *inside* the stopped improvement loop.
  Fleet census 2026-08-03: **204 `detected` across 10 sites against 2 `triaged`**.
  `validate_page_content.go:644-650` already refused to file work items for exactly
  this reason, in writing. IMP-016 states the rule
  (`register/improvement-loop.md:130-136`).
- **Candidates 3 and 4 reverse an owner ruling** (`bugs_open/136:32-35`, 2026-07-29)
  and are gated as **G1**, an explicit separate owner go.
- **Candidate 2 is warned against by the 149 lane** (`149:395-398`) as an
  unattributable change until B2 is settled.

So the ordering constraint is the reverse of what the file assumed: **detection
cannot be widened until the promotion gap (`bugs_open/083`) is answered.** Writing
the seam anyway would be a textbook case of the thing this estate keeps relearning
— shipping a mechanism that reads as coverage and drains nowhere.

Handed back OPEN with the record corrected. No code written, deliberately.

## 2026-08-04 (evening) — the diagnosis verdict, read at last: UNVERIFIABLE

Run corr `54bf4506-…` completed same evening it was filed: **UNVERIFIABLE —
"Diagnosis NOT confirmed (stopped: scope-not-narrowing). Best-effort trail attached
for a human; no fix proposed."** Four bundles, no refutation, no corroboration.

So the mechanism claim in the bug file's STATUS block stands on the first-hand
verification recorded in this file (the scheduler's `loadDueTasks` read directly,
the live `scheduled_tasks` rows, the workflow steps, the `last_audit` stamps) — the
declared substitute the 2026-07-31 owner ruling permits, now declared **with** the
loop's non-answer attached rather than silently. An UNVERIFIABLE is not a REFUTED:
nothing in the trail contradicts the mechanism, and nothing independently confirms
it either.
