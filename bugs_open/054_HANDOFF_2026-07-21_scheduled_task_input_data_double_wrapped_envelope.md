# 054 — scheduled-task `input_data` authored as a full message envelope, so the payload double-wraps and never reaches the action

> **VERIFIED & LIVE 2026-07-21.** After the data fix, a forced run
> (`9964b72e-341b-4693-b181-f03e3264ef0f`) reached `directory_export_json` with
> `domain: "vetcomparison.uk"` and **built 5 files** — directory of 2,109 businesses,
> 13 price-aggregate rows — then sent them to the git-adapter
> (`export_json.git_result.file_count = 5`), reaching step `complete` with no error.
> That is behaviour an empty domain could never produce; the config now demonstrably
> drives the export. The live site's `directory-metadata.json exported_at` advances past
> 2026-07-17 once the git-adapter commit lands and the site redeploys (downstream, no
> further action needed). **This is a DB-config + seed fix — no image roll required.**

**Filed 2026-07-21.** **Status: FIXED (data live; seed committed).** Fleet-wide authoring trap; one live casualty (`directory-export-json`, now fixed).

This is the **sibling case of `bugs_closed/042`** (correlation `55dc0fa4-116c-40d6-90b2-bfad9ad73692`),
opened as its own case per 042's own instruction ("if it needs a home, open a new case rather than
reopening this one").

## 042 mis-attributed this — the correction is the point

042's §Related and its fix-candidate 2 describe this sibling as *"a literal domain string does not
reach the action"* — i.e. the same `ExtractActionInputs` literal-string family as 042 itself. **That
is wrong, and it was never checked against the code.** `DirectoryExportJSONAction` does **not** use
`ExtractActionInputs` at all — it reads `config["domain"].(string)` directly
(`directory_export_action.go:71`) after merging `CollectedData["input_data"]` into config
(`:127-131`). The domain is a perfectly ordinary string; the problem is that it is **one nesting
level too deep**, not that a literal string is unreadable. Fixing the literal-string half of
`ExtractActionInputs` would do nothing for this bug.

> **Transferable lesson:** a family label ("literal not reaching an action") grouped two bugs with
> different mechanisms. The grouping was written from the symptom (domain absent → abort) without
> reading the failing action. Read the action before assigning a case to a family.

## One-line

The scheduler's `fireTrigger` **always** wraps a task's `input_data` in its own envelope
`{action, config:{agent_type}, input_data:<the column>}` (`cmd/scheduler/main.go:396-402`). So
`scheduled_tasks.input_data` must contain **only the payload**. `directory-export-json` was seeded
with a *whole message envelope* as its payload, so the real fields end up at
`input_data.input_data.*` and the action, which reads one level down, sees no `domain`.

## Mechanism — traced end to end against the live failed run (`6271b72d-...`)

1. **DB.** `scheduled_tasks.input_data` for `directory-export-json` =
   ```json
   {"action":"orchestrate","config":{"agent_type":"directory-json-exporter"},
    "input_data":{"domain":"vetcomparison.uk","vertical":"veterinary", ...}}
   ```
   Authored by `docs/.../vetcomparison/009_directory_export_agents.sql:103-111`
   ("Modelled on the med-json-exporter pair (037)" — so 037/`med-export-json` shares the defect).

2. **Scheduler.** `fireTrigger` (`cmd/scheduler/main.go:396-402`) builds the Kafka body as
   `{action:orchestrate, config:{agent_type}, input_data: <the entire column above>}`. The column is
   already an envelope, so the body is now **double-enveloped**: the payload sits at
   `body.input_data.input_data.domain`.

3. **Chassis.** `BuildCollectedData` (`datahelpers/data_helpers.go:873-876`) sets
   `collectedData["input_data"] = unnestedBody["input_data"]` =
   `{action, config, input_data:{domain...}}`. Confirmed verbatim in the failed run's
   `collected_data->'input_data'`.

4. **Action.** `DirectoryExportJSONAction` merges `collectedData["input_data"]`'s keys into config
   (`:127-131`) — it gets `action`, `config`, `input_data`, **but no `domain`**. `ec.Domain == ""`
   → `return nil, fmt.Errorf("directory export requires an explicit domain; refusing ...")`
   (`:134-136`). This is **correct fail-closed behaviour** — the action refused rather than exporting
   under a wrong/empty domain.

Live evidence: `orchestration_states` row `6271b72d-fb93-4af9-864c-59957ec75a13`, owner
`business-intel`, `status=FAILED`, `current_step=export_json`, error
`... directory_export_json: directory export requires an explicit domain`, `created_at`
2026-07-19 20:25:50 (matches the task's `last_triggered_at`).

## Blast radius

- **Live casualty:** vetcomparison.uk's business directory. `directory-metadata.json` on the live
  site shows `exported_at: 2026-07-17T20:23:37Z` — every scheduled refresh since has hit this abort.
  (The 07-17 export predates the enable and came via a different path.)
- **Same-family, currently harmless:** `diagnose-pipeline-trigger` (enabled) is authored with the
  same envelope but an **empty** payload `input_data:{}`, so there is nothing to bury — the
  diagnose loop reads no payload fields. It "works" only by that accident.
- **Same-family, disabled:** `med-export-json` (`{...,"input_data":{"domain":""}}`) and
  `med-discover-urls` (`{...,"input_data":{}}`) — vet-med-export workstream. They will fail the same
  way if enabled with a real payload. **Recommend that workstream apply the identical unwrap** to
  the live rows and to seed `037_vet_med_export_orchestrator_prices_json.sql:145-151`. Not fixed here
  to stay within scope / not touch another workstream's live state.

## The contract (so nobody re-authors the trap)

`scheduled_tasks.input_data` is the **payload only**. `fireTrigger` supplies `action`,
`config.agent_type` and the `input_data` wrapper itself. Correct examples already in the table:
`ch-enrichment` = `{batch_size, vertical_slug}`, `vet-batch-verify` =
`{task_type, batch_size, vertical_slug}`, `vet-sweep-continue` = `{limit, country, ...}`. Never put
`action`/`config`/`input_data` keys **inside** `scheduled_tasks.input_data`.

## Fix applied

- **Data (live, immediate — DB config, no image roll):**
  ```sql
  UPDATE scheduled_tasks SET input_data = input_data->'input_data', updated_at = NOW()
  WHERE name = 'directory-export-json';
  ```
  Takes the buried inner payload and promotes it to the top level, exactly. Then force a run:
  `UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'directory-export-json';`
- **Seed (committed, so a re-seed can't reintroduce it):**
  `009_directory_export_agents.sql` — `input_data` reduced to the payload, with a comment recording
  that `fireTrigger` supplies the envelope.

No code change: the action's fail-closed refusal is the behaviour we want. Adding "smart" unwrapping
to `fireTrigger` (detect a nested `input_data` and flatten it) would reintroduce exactly the
regression class 042 warned about — a legitimate payload field named `input_data`/`action`/`config`
would be silently eaten. The mistake is in the data; fix the data.

## How to verify (do NOT re-read the config)

Set the data as above, force the fire, then watch behaviour:
```sql
-- a fresh run must reach 'complete', not FAIL at export_json:
SELECT status, current_step, left(error,80)
FROM orchestration_states
WHERE orchestration_name LIKE 'sched-directory-export-json-%'
ORDER BY created_at DESC LIMIT 1;
```
Then the artefact, not the status: `curl -s https://vetcomparison.uk/data/directory-metadata.json`
— `exported_at` must advance past 2026-07-17 once the git-adapter commit lands and the site
redeploys.

## Related

- `bugs_closed/042` — the numeric-scalar sibling (FIXED & LIVE v1.0.1144). Same symptom class
  ("config value never reaches the action"), different mechanism.
- `037_vet_med_export_orchestrator_prices_json.sql` — the seed 009 was modelled on; carries the
  same defect for the (disabled) med exporter.
