# 344 — a workflow's OWN `config` is a searchable candidate source, so any step-config key can be injected as an input value

**Filed** 2026-08-21 by the `staged_component_build` lane (RFC_029 step 5 census).
**Found by** dispositioning the smallest conflict class on the board — 3 rows — and reading where
its winner came from. The conflict rows are trivial; **the silent population behind them is not.**
**Status** OPEN. Mechanism confirmed first-hand; exposure measured; damage NOT yet demonstrated.

## 1. The observation that started it

`generic` / `summary`, 3 conflict rows on 2026-08-17, winner:

```
config.workflow.steps.probe_control.config.summary
```
with all 3 candidates of the form `config.workflow.steps.probe_*.config.summary`.

The resolver descended into the orchestration's **own workflow definition** and offered a *step's
configured `summary` string* as the value for a requested `summary` input.

## 2. Mechanism

`findFieldRecursive` walks `collected_data` and skips a fixed set of infrastructure keys
(`isInfrastructureKey`, `unified_extractor.go:720-745`): `agent_config`, `__raw_message__`,
`__work_request__`, `__execution_context__`, the topic keys, and `retry_payload`.

**Plain `config` is not in that list.** So when `collected_data` carries a `config` key, everything
under it — up to the depth cap — is a legitimate candidate for any requested field name.

The `retry_payload` entry's own justification applies here almost word for word: *"nothing resolves
fields OUT of it, so the search must not treat its contents as live data."* A workflow definition is
the same kind of thing: it is the instructions, not the data.

## 3. Exposure, measured

```sql
SELECT count(*) AS orchestrations,
       count(*) FILTER (WHERE collected_data ? 'config') AS carries_config
FROM orchestration_states;
```
→ **1941 of 3107 (62.5%)** carry a `config` key.

Most are harmless — `config` holds only `{agent_type}` for `endpoint-health-checker` (965),
`build-pipeline-trigger` (313), `availability-discovery-agent` (150), `council-gate` (63). The
exposure is the rest:

```sql
SELECT owner_agent_type, (SELECT string_agg(k,',' ORDER BY k) FROM jsonb_object_keys(collected_data->'config') k), count(*)
FROM orchestration_states WHERE collected_data ? 'config' GROUP BY 1,2 ORDER BY 3 DESC;
```
→ **208 `generic` orchestrations carry `config.workflow`** — the *entire* workflow definition, and
therefore **every step's config, for every step**, as candidate values.

So on those runs, any requested field whose name collides with a config key anywhere in the
workflow (`summary`, `description`, `reason`, `page_id`, `mode`, `page_name`, …) can be resolved
**from the workflow's own text**.

## 4. Why the conflict count badly understates it `[INFERRED — not yet measured]`

A conflict row requires the collected candidates to **differ**. The 3 rows exist only because that
probe workflow happened to have three steps with three different `summary` strings. **A workflow
with exactly one matching config key substitutes silently** — no WARN, no row — and the value is
still the wrong kind of thing.

This is the same silent half as `bugs_open/330` §4, on a different surface. Measuring it needs the
stripped-probe method the 330 lane built (330 §9 and the RUNBOOK), applied with `config` included
rather than excluded.

## 5. What it is NOT

- **Not `agent_config`.** That key is already skipped. This is plain `config`.
- **Not step 5's blocker.** With 3 conflict rows this pair is dispositioned for the flip either way:
  under Phase 2 it resolves to nothing, which is correct. **The flip does not fix the silent half**,
  which is why this is a separate bug and not a line in the step-5 census.
- **Not demonstrated damage.** I have the mechanism and the exposure. I do **not** have a case where
  a wrong value from `config` changed an artefact. Do not upgrade this without one.

## 6. Fix candidates, ordered by what closes the door

1. **Add `config` to `isInfrastructureKey`.** Four characters, closes both halves — conflicting and
   silent — for every field and every agent, and it matches the stated reason the list already
   exists. Blast radius must be measured first: any step that *legitimately* resolves a field out of
   `config` today would stop. Query for that before building — a direct request FOR the key itself
   still resolves, as the `agent_config` comment notes, so only recursion is affected.
2. **Stop putting the whole workflow in `collected_data`.** Narrower blast radius conceptually but a
   bigger change, and it is somebody else's surface — the 208 rows are `generic`, so find the
   producer first.
3. **Per-field `?` markers.** Correct but endless: this is a property of the search, not of any one
   field, so fixing it field-by-field is the wrong altitude.

## 7. How to verify a fix

For candidate 1, the mutation proof is direct: a tree carrying `config.workflow.steps.X.config.<f>`
and nothing else named `<f>` must resolve `<f>` to **nothing** after the change and to the config
string before it. That is a unit test on `findFieldRecursive`, not a runtime observation — no live
demand needed, which is a rare luxury in this family.

## 8. Related

- `bugs_open/330` — the same silent-substitution class via a different route (declared-but-empty
  falling through to the search). Its §9 audit method is reusable here.
- `bugs_closed/306` — the tie-break that decides which candidate wins once several exist.
- RFC_029 §9 D2 — the Phase 2 flip, which closes the conflicting half of this and not the silent half.
