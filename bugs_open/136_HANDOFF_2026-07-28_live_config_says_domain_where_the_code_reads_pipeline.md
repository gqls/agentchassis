# 136 — Live config says `*_domain` where the code reads `*_pipeline`, and nine step-config keys are read by nobody

**Filed** 2026-07-28 (late evening) · **Status** OPEN, unowned · **Class** `bugs_closed/101`
(unknown config key silently ignored) · **Found by** the first systematic pass over pile B of
the config-key ratchet (`bugfix_100_101_scrape_provenance/HANDOFF_2026-07-28b` §3b)

---

## The one-line version

Three actions were renamed internally from **`domain`** to **`pipeline`**, the live
`agent_definitions` kept the old word, and **nothing in the fleet sets the new one** — so on
every one of those steps the config is inert and the action silently uses its hardcoded
default. In all three cases the default currently *happens* to equal what the config asks
for, so nothing looks broken. That coincidence is the bug: the config reads as a
specification of behaviour while being evidence of nothing.

A separate key on the same sweep, `create_work_item.summary_template`, is **not** latent —
it is producing wrong data in the human-review queue today.

---

## 1. How this was found (so it can be re-run)

The ratchet's pile B is "the action registered an `ActionInputSpec`, but a live step carries a
key that is in neither `Required`, `Optional`, `ConfigKeys` nor `Deprecated`". 34 actions were
in it. Rather than read all 34, each missing key was first grepped against its own action's
source file; the eight that appeared **zero** times there were then read properly. Six of the
eight were dead. Re-derive the pile with:

```bash
./scripts/audit-config-keys.sh --json > gap.json
go run ./cmd/config-key-audit --specs   > specs.json
# join: for each undeclared action, live keys minus (required ∪ optional ∪ config_keys ∪ deprecated ∪ framework)
```

**The grep is a ranking heuristic, not a verdict.** Two escape hatches make a literal absent
from the source and still live, and both were checked before any key below was called dead:

- `datahelpers.BuildDeprecationMap` materialises `<field>_field` aliases into `spec.Deprecated`
  — but the audit already counts `Deprecated` as known, so a key it covers is never flagged.
- `ExtractActionInputs` **Strategy 1** (`action_inputs.go:386-401`) populates `Values` from
  whatever `config["input_fields"]` names, which *can* be outside the spec. **Strategy 0 and
  Strategy 2 iterate `Required ∪ Optional` only** (`:359-360`, `:368`). So a key outside the
  spec is readable **only** if that step also sets `input_fields`. Every claim below was
  checked against the step's actual `input_fields`.

---

## 2. The `*_domain` / `*_pipeline` split — the dominant sub-family

| key in live config | live steps | key the code actually reads | live steps setting the code's key | default when unset |
|---|---|---|---|---|
| `check_domain` | 3 | `check_pipeline` | **0** | `"design"` |
| `target_domain` | 3 | `target_pipeline` | **0** | `"build"` |
| `item_domain` | 7 | `item_domain` ✔ read | — | — |
| `item_pipeline` | 2 | `item_pipeline` ✔ read | — | — |

Measured 2026-07-28 against live `agent_definitions` (active, non-snapshot, not deleted):

```sql
SELECT ck.key, count(*) AS live_steps
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v),
     jsonb_object_keys(e.v->'config') AS ck(key)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND jsonb_typeof(e.v->'config')='object'
  AND ck.key IN ('target_pipeline','target_domain','check_pipeline','check_domain',
                 'item_pipeline','item_domain')
GROUP BY 1 ORDER BY 1;
--  check_domain  | 3
--  item_domain   | 7
--  item_pipeline | 2
--  target_domain | 3
```

**`item_*` is fine — both names are read.** The two broken pairs share a tell worth naming:
*the key the code reads is set by zero definitions fleet-wide.* A key nothing sets, whose
default the code always takes, is the signature of a rename that landed on one side only.

### 2a. `run_discovery_checks.check_domain` — 3 agents, 3 different values, all discarded

`discovery_checks.go:66-69` reads `config["check_pipeline"]` and defaults it to `"design"`.
`config["check_domain"]` is read nowhere in the tree (`grep -rn check_domain --include=*.go`
returns nothing). `registry.go:46` states the intent outright:
`Pipeline string // check_pipeline from config, e.g. "design"`.

The three live carriers each ask for something different:

| agent | `check_domain` | `check_pipeline` |
|---|---|---|
| `design-discovery-agent` | `design` | *(unset)* |
| `completeness-discovery-agent` | `content` | *(unset)* |
| `quality-discovery-agent` | `build` | *(unset)* |

`dctx.Pipeline` reaches `site_work_items.pipeline` via `WorkItemSpec.Pipeline`
(`discovery_checks.go:143-151`). **[MEASURED] Nothing is mislabelled today.** Only six call
sites propagate `dctx.Pipeline`, and the four checks among them — `palette_contrast`,
`image_url_404`, `unfulfilled_image_prompt`, `placeholder_image_in_use` — appear *only* in
`design-discovery-agent`'s check list, where the `"design"` default is the right answer by
luck. Confirmed against the data rather than inferred:

```sql
SELECT created_by, pipeline, count(*) FROM site_work_items
WHERE source='discovery' GROUP BY 1,2 ORDER BY 1;
-- completeness-discovery-agent | build   | 88     (no 'design' rows)
-- completeness-discovery-agent | content | 56
-- quality-discovery-agent      | build   |  7     (no 'design' rows)
```

Neither non-design agent has produced a single `design`-pipeline row. Most checks hardcode
their own pipeline in the `WorkItemSpec` they return, which is what keeps this contained.

**Why it is still a defect:** the containment is a coincidence of today's check-to-agent
mapping. The moment a check that propagates `dctx.Pipeline` is added to the content or build
agent's list — or a design check moves — those findings are written as `design` and land in
the wrong queue, with no error and nothing in the config to suggest why.

### 2b. `triage_detected_items.target_domain` — 3 agents, defaults to the same value

`triage_detect_items_action.go:83-86` sets `targetPipeline := "build"` then overrides from
`config["target_pipeline"]`. `target_domain` is read nowhere. All three carriers
(`site-review-agent`, `improvement-loop`, `design-audit-agent`) set `target_domain: "build"`
— **identical to the default**, so behaviour is correct today and the key has never once
mattered. Change any of them to `content` and nothing happens, silently.

---

## 3. The other dead keys from the same sweep

| action | key | carrier | code reads instead | consequence |
|---|---|---|---|---|
| `create_work_item` | `summary_template` | `grounded-explainer` | `summary` (`:133-136`) | **BITING — see §4** |
| `create_work_item` | `spec_fields` | `grounded-explainer` | `spec_data` / `spec_paths` / `spec_literal` (`:190-207`) | the four named fields never reach the item spec |
| `create_work_item` | `domain` | `claims-auditor` | `inputs.Get("domain")`, not in spec, step sets no `input_fields` | `item_key` falls back to `siteID[:8]` (`:147-155`). **[UNEXERCISED]** — `item_key LIKE 'claims_llm%'` returns **0 rows**, so this has never run |
| `prepare_training_data` | `s3_bucket` | `training-data-preparer` | `bucket` | **doubly dead** — see below |
| `prepare_training_data` | `input_mapping` | `training-data-preparer` | nothing | inert |
| `fix_component_template` | `fix_type_field` | `component-template-fixer` | `fix_type` (`:117`) | harmless: intent is served instead by `inferFixTypeFromCategory` (`:126`) |
| `plan_sections` | `domain` | `page-build-handler` | spec has `pipeline`, not `domain` | inert |

**`prepare_training_data.s3_bucket` deserves its own line.** The action reads `bucket`, not
`s3_bucket` — so the key is dead. *And* its value names a bucket that does not exist. This was
already discovered independently and written into the tree at
`internal/adapters/thunder/adapter.go:199-204`:

> `"personae-model-training"` is the REAL bucket holding training data (confirmed against the
> live B2 bucket list 2026-05-23). NOTE: the `training-data-preparer` agent_def lists
> `s3_bucket="finetuning"`, but no such bucket exists — that value is stale/logical.

Someone hit the wrong-bucket half months ago and worked around it in the adapter. Nobody
noticed the key was never read either way.

---

## 4. The one that is biting: `create_work_item.summary_template`

`grounded-explainer`'s `create_review_item` step sets:

```json
{ "item_type": "grounded_draft_review",
  "summary_template": "Grounded explainer draft ready for review: {{.input_data.topic}}",
  "spec_fields": ["draft","grounding_audit","registration","input_data"],
  "handler_agent": "human-review", "status": "needs_human_review" }
```

It sets **no `summary`**. `create_work_item_action.go:133-139`:

```go
summary := inputs.Get("summary")
if summary == "" { summary, _ = config["summary"].(string) }
if summary == "" { summary = itemType }
```

So the summary falls through to the *item_type*. Confirmed in live data:

```sql
SELECT created_by, summary, count(*) FROM site_work_items
WHERE created_by ILIKE '%grounded%' GROUP BY 1,2;
-- grounded-explainer | grounded_draft_review | 2
```

Two items are sitting in the human-review queue captioned **`grounded_draft_review`** instead
of *"Grounded explainer draft ready for review: <topic>"*. The reviewer sees a type name.
**The fallback at `:137-139` is what hides it** — without it the empty summary would have been
noticed immediately.

Note `summary_template` also implies templating (`{{.input_data.topic}}`) that
`create_work_item` does not do at all for summaries. Implementing the key means implementing
the render, not just renaming it.

---

## 5. Fix candidates, ordered by what closes the door

1. **Make the bad state unrepresentable — opt these actions into the detector.** All nine keys
   above are invisible precisely because their actions have not opted into unknown-config-key
   detection. `run_discovery_checks`, `triage_detected_items`, `create_work_item`,
   `prepare_training_data`, `fix_component_template` and `plan_sections` should declare their
   contract so the *next* rename is caught at validation instead of months later.
   **Prerequisite, and it is not optional:** several of these have genuine **spec gaps** that
   must be closed first or opting in would emit false warnings — `run_discovery_checks` reads
   `checks` (`:73`) and `check_pipeline` (`:66`) and its spec declares neither. Add what the
   action really reads, then set `CheckConfig: true`.
2. **Fix the definitions** — rename `check_domain`→`check_pipeline`,
   `target_domain`→`target_pipeline`, `summary_template`→`summary`, `s3_bucket`→`bucket`,
   `fix_type_field`→`fix_type`; delete `spec_fields`, `input_mapping`, `plan_sections.domain`.
   Cheap, but on its own it fixes today's instances and prevents nothing.
3. **`summary_template` needs a decision, not a rename** — either implement template rendering
   in `create_work_item`, or change the definition to a literal `summary`. Renaming the key
   without rendering gives the reviewer a raw `{{.input_data.topic}}` string.
4. **Do NOT "fix" any of these by adding the key to `ConfigKeys`.** Declaring a key the action
   does not read makes it *recognised*, silences the detector, and leaves the behaviour broken
   — the recorded `WRONG_CALLS.md` 2026-07-28 mistake, committed by the fix for `101` itself.

---

## 6. Landmines

- **The defaults hide it.** Both `*_domain` cases behave correctly today because the hardcoded
  default equals what the config asks for. Any test asserting current behaviour passes. There
  is no error, no warning and no failed run anywhere in this bug.
- **`grep <key> --include=*.go` returning nothing is a lead, not a verdict** — check whether
  the step sets `input_fields` (Strategy 1 reads outside the spec) before calling a key dead.
- **A zero row count means unexercised, not clean.** `claims_llm%` returned 0 rows; that is why
  §3's `domain` row is marked `[UNEXERCISED]` rather than refuted.
- **`create_work_item`'s summary fallback (`:137-139`) is the reason this survived.** A defaulting
  fallback over a missing config value converts a loud failure into a plausible-looking wrong
  value.

## 7. Verify after any fix

```bash
./scripts/audit-config-keys.sh          # these keys must leave the UNDECLARED section
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=-1 --since=1h \
  | grep "keys this action does not read"     # -l, NOT logs deploy/… (reads one pod of N)
```

```sql
-- the summary must stop equalling the item_type
SELECT summary, item_type FROM site_work_items WHERE created_by='grounded-explainer';
```
