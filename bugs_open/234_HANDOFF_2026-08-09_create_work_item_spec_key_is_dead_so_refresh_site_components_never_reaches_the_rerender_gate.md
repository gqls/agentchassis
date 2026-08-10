# 234 — `create_work_item`'s `spec` config key is read by nothing, so improvement-loop's `refresh_site_components` has never reached the rerender gate

**Filed 2026-08-09** from the `bugfix_136_config_key_aliases` lane, found by the no-op check
before opting `create_work_item` into unknown-config-key detection (`bugs_open/136` §12).

**Status: OPEN — IN PROGRESS (bugfix_234_dead_spec_key lane, opened 2026-08-10). The owner
decision below is TAKEN: restore the flag. Fix underway per
`docs/agent_docs/docs024_key_docs_latest/bugfix_234_dead_spec_key/PLAN_2026-08-10_dead_spec_key.md`.**

> **CORRECTED 2026-08-10 (bugfix_234 lane):** two premises of the owner-decision framing
> below went stale within a day of filing, and both moved the decision.
> 1. *"restore it behind whatever guard 226 lands"* — **226's guard HAS landed**: both
>    chassis replicas (v1.0.1274 at check time) carry `emitChromeDivergenceItem`
>    (strings-proven, negative control clean), 226 is "DONE IN SUBSTANCE" and its guard has
>    already refereed a real event. The interaction this file deferred on is closed.
> 2. *"turning on full site-component reassembly that has not run from this path in
>    months"* — true of THIS path, but the behaviour is not dormant fleet-wide: **8
>    producers file `refresh_site_components: true` at ~5–15 rows/day** (measured
>    2026-08-09). Restoring this path adds ~1.8 rows/day to a routine daily behaviour.
>
> **OWNER DECISIONS (2026-08-09/10, at plan approval):** (a) candidate 2 resolves to
> **restore the flag via `spec_literal`**; (b) close the class BOTH ways — `StrictConfig:
> true` on `create_work_item` (its doc comment's precondition is now met: recognised set
> verified complete against every live step at ALL depths, 2026-08-09) AND a new opt-in
> `ActionInputSpec.RemovedConfigKeys` field (retired key → hard validation error naming
> the replacement). Candidate 3's prohibition stands: `spec` is still never declared in
> `ConfigKeys`.

---

## The one-line version

Three live steps set a `create_work_item` config key called **`spec`**. The action has never
read a key by that name — it builds the item spec from `spec_data` / `spec_paths` /
`spec_literal`. So every item those steps file carries `spec = {}`, and for
`improvement-loop.insert_rerender_item` that silently drops
`refresh_site_components: true` on the floor, which is the flag the rerender path gates the
site-component reassembly on.

## The evidence

**The action does not read it.** `create_work_item_action.go` composes the spec in three
layers at `:208-236` — `spec_data` (a path to a map), `spec_paths` (`{key: path}`),
`spec_literal` (`{key: value}`). There is no `config["spec"]` and no
`inputs.Get("spec")`/`GetMap("spec")` anywhere in the file.

**Nothing else could resolve it either**, which is the part that needs reading rather than
grepping (this is `bugs_open/136` §6's landmine). Every strategy in `ExtractActionInputs`
iterates one of three sets: `spec.Required ∪ spec.Optional` (Strategies 0, 2, 4 and the
nested-object pass), `config["input_fields"]` (Strategy 1), or `spec.Deprecated`
(Strategy 3). `spec` is in none of them, and none of the three steps sets `input_fields`.

**The three carriers** (live, 2026-08-09, walked at all depths):

| agent | step | value it intends |
|---|---|---|
| `improvement-loop` | `insert_rerender_item` | `{"refresh_site_components": true}` |
| `improvement-loop` | `record_not_converging` | `{"reason": "…fixes are not landing…", "capability": "audit_not_converging"}` |
| `deduplicate-sections` | `queue_rerender` | `{"page_id": "input_data.page_id"}` |

**The consumer chain that never receives it** — and it is config, not Go, which is why a
symbol search will not find it:

- `051_build_dispatch_loop.sql:823` — input_mapping carries
  `"refresh_site_components": "pending.first_item.spec.refresh_site_components"`
- `033_rerender_pages_action.sql:1107` — the conditional is
  `"input_data.spec.refresh_site_components == true OR input_data.refresh_site_components == true"`

**The damage, measured:** every `improvement_rerender_*` row ever filed carries an empty
spec — **16 of 16**, the most recent on the day of filing. Positive control: **4,972** rows
fleet-wide DO carry a non-empty spec, so the column and the mechanism work fine.

```sql
SELECT count(*) AS rows, count(*) FILTER (WHERE spec = '{}'::jsonb) AS empty
FROM site_work_items WHERE item_key LIKE 'improvement_rerender%';   -- 16 | 16
```

The other two steps have **never filed a row** (`dedupe_rerender%` → 0,
`capability_gap_audit%` → 0), so their intent has never been exercised either way.

## Root cause

`bugs_closed/024` introduced `spec_paths` and `spec_literal` (migration 180, 2026-07-21)
precisely because `spec_data` alone could not express a constant. **These three steps were
never migrated onto the new spelling** — they kept a `spec` key that no version of the
action has read. It is the same family as `bugs_open/136` (config says X, code reads Y) and
the same shape as its §4 `summary_template`: a key that reads as wired, with a defaulting
path underneath that turns the failure into a plausible-looking empty value.

A council seat was shown this exact config in a 2026-07 round on bug 024 — the improvement-
loop step config with `"spec": {"refresh_site_components": true}` is quoted verbatim in the
verdict note in `doc_notes` — and nobody noticed the key was dead. It is not obvious by
reading; it is only obvious by comparing against what the action reads.

## Fix candidates, ordered by what closes the door

1. **Translate the two zero-exposure steps now** — behaviour-restoring and unexercised:
   `deduplicate-sections.queue_rerender` `spec` → **`spec_paths`**
   (`{"page_id": "input_data.page_id"}` is a path), and
   `improvement-loop.record_not_converging` `spec` → **`spec_literal`** (both values are
   constants). Neither step has ever filed a row, so there is no live behaviour to change
   and no existing row to make inconsistent. Data-only, live immediately, seeds in the same
   commit (`054_improvement_loop.sql:166`, `269_deduplicate_sections_handler.sql`).
2. **`improvement-loop.insert_rerender_item` needs an OWNER DECISION, not a migration.**
   Translating it to `spec_literal` would start delivering `refresh_site_components: true`
   on ~2 rows/day, turning on full site-component reassembly that has not run from this
   path in months. **That interacts with `bugs_open/226` (chrome rebuild silently discards
   hand-patched content, no divergence warning).** The choice is genuinely three-way:
   restore the flag; delete the key and accept that this path never refreshes chrome;
   or restore it behind whatever guard 226 lands. Do not pick one while tidying something
   else — that is how the flag got lost in the first place.
3. **Do NOT declare `spec` in `CreateWorkItemInputSpec.ConfigKeys`.** It would make the key
   recognised, silence the detector, and leave the behaviour broken — the `bugs_closed/101`
   failure mode, and explicitly forbidden by `bugs_open/136` §5.4. `create_work_item` is now
   opted into detection and **reports this key on purpose**; the audit's `UNKNOWN KEYS`
   count reading 1 (not 0) after the next chassis roll is this bug being visible, not a
   regression.

## Diagnosis-loop verdict: UNVERIFIABLE — and why that is not a refutation

Filed to the loop before this file was written, per the 2026-07-31 owner ruling
(`090` dispatch correlation **`be967639-d195-444a-b9c3-ef1445ff7ae1`**, completed
2026-08-09). Outcome: **UNVERIFIABLE**, `runtime_site: build-dispatch-loop`.

An UNVERIFIABLE says the loop could not get the evidence — it does not say the premise was
false (`an-unverifiable-verdict-does-not-say-your-premise-was-false`). Its own
`needed_evidence` names why, and both reasons are about its tooling rather than this
hypothesis:

- **Its code index is ~2 days stale** (it says so), so searches for `insert_rerender_item`,
  `record_not_converging` and `RefreshSiteComponents` returned 0 rows — which it correctly
  recorded as "unknown, not absent". The last of those would have returned 0 anyway: the
  consumer is **workflow config in a seed**, not a Go symbol.
- **Its own `data_requests` were composed against a schema that does not exist.** Both read
  `SELECT agent_type, step_name, config FROM agent_definitions …`. That table has `type` and
  a `default_config` jsonb; **none** of `agent_type`, `step_name`, `config` is a column
  (`information_schema.columns` → 0 of 3). Those queries could never have returned a row, so
  the step configs it needed were unreachable by construction.

**Substituted first-hand verification, stated plainly as the ruling requires:** the action
body read end to end (no reader of `spec`); every `ExtractActionInputs` strategy read and
enumerated; the three live step configs read from the DB at all depths; 16/16 empty specs
measured with a 4,972-row positive control; both consumers read in the seed files at the
line numbers cited above.

**Transferable, and worth more than this bug:** any diagnosis whose evidence lives in
`agent_definitions` **step configs** will come back UNVERIFIABLE for the same reason, because
the diagnoser reaches for columns that do not exist. That is a defect in the loop's SQL
authoring, not in the hypotheses it is given. Added to `016b` §9.

## How to verify a fix

```sql
-- after translating a step, the filed row must carry the value
SELECT item_key, spec FROM site_work_items
WHERE created_by='improvement-loop' AND item_key LIKE 'improvement_rerender%'
ORDER BY created_at DESC LIMIT 3;   -- spec must contain refresh_site_components, not {}
```
Do not verify at the definition — a config that *looks* right is exactly what this bug is.
Verify at a **filed row**, and make sure the row was filed after the change.

## Related

- `bugs_open/136` §12 — where this was found, and why it is deliberately not fixed there
- `bugs_closed/024` — introduced `spec_paths`/`spec_literal`; these steps never migrated
- `bugs_open/226` — the chrome-rebuild risk that makes candidate 2 an owner call
- `bugs_closed/101` — why declaring the key to clear the audit would be the wrong fix
