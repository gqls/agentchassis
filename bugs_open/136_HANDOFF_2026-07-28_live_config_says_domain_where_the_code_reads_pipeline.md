# 136 — Live config says `*_domain` where the code reads `*_pipeline`, and nine step-config keys are read by nobody

**Filed** 2026-07-28 (late evening) · **Status** OPEN · **Class** `bugs_closed/101`
(unknown config key silently ignored) · **Found by** the first systematic pass over pile B of
the config-key ratchet (`bugfix_100_101_scrape_provenance/HANDOFF_2026-07-28b` §3b)

> **UPDATE 2026-07-29 — fix candidate 1 is APPLIED for the two `*_domain` actions, and the
> defect is now REPORTED instead of silent.** Commit `099476b56`: `run_discovery_checks` and
> `triage_detected_items` now declare exactly the keys they read
> (`checks`+`check_pipeline`, and `target_pipeline`), which opts both into the detector. The
> spec gaps flagged as a prerequisite in §5 were closed in the same commit — that was the
> whole reason they were a prerequisite. **`scripts/audit-config-keys.sh` UNKNOWN KEYS went
> from `none` to naming exactly `check_domain` and `target_domain`**; ratchet 152→150
> undeclared, 58→60 declared. Council corr `1c606c72-eb82-4761-b30d-1a7c653b744d`.
>
> **Behaviour is unchanged, verified rather than argued:** `ConfigKeys` is read only by
> `UnknownConfigKeys`, `ListDeclaredConfigKeys` and `cmd/config-key-audit` —
> `ExtractActionInputs` never reads it (`grep -rn '\.ConfigKeys' platform/ cmd/`).
> `StrictConfig` stays unset, so it warns and blocks nothing.
>
> **LIVE on v1.0.1197** — established by ancestry bracketing, not by the tag: a 07:47Z
> commit's string is present in the binary, a 07:58Z commit's is absent, negative control
> clean, and this commit is 07:16:56Z. (A direct pod-grep was impossible: every string
> literal the change adds already existed, because the actions read those keys.)
>
> **A THIRD INSTANCE was then found and opted in — `plan_sections.domain`** (`07340d1e2`,
> council `30a8785b-8cad-4d10-8633-486d81e837e9`). See §2c. The audit now names all three.
>
> **Still OPEN**, on two counts: the definitions are unfixed (§5, other lanes' agents), and
> the runtime warning has **still not been witnessed in a pod**.

> **OWNER RULING 2026-07-29 — the improvement loop is stopped DELIBERATELY**, during a heavy
> development phase. `improvement-sweep` (`enabled=f` since 2026-05-02) and the discovery
> one-shot are a **decision, not a defect** — do not re-file them as dead scheduled tasks and
> do not re-enable them.
>
> **This is why §2a/§2b's warning cannot be the witness.** The six agents carrying
> `check_domain`/`target_domain` are quiesced *by design*: `run_discovery_checks` last ran
> 17h before the roll, `triage_detected_items` not in 24h, and no enabled `scheduled_tasks`
> row targets any of them. Firing one would push real work items into a queue deliberately
> stilled — **a verification that requires restarting something the owner stopped on purpose
> is not a verification worth having.** Hence §2c.

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

### Why `item_*` is fine is the most important line in this bug (added 2026-07-29)

`create_work_item` handles **this exact rename** with an explicit back-compat shim —
`create_work_item_action.go:118-121`:

```go
itemPipeline, _ := config["item_pipeline"].(string)
if itemPipeline == "" {
    itemPipeline, _ = config["item_domain"].(string) // backwards compat
}
```

So this was **not** an unnoticed inconsistency in naming. The `domain` → `pipeline` migration
was known, was deliberate, and someone wrote the fallback that makes old definitions keep
working. **They wrote it on one action and not on the other two.** That reframes the bug from
"two places spell it differently" to "a migration was applied inconsistently, and the two
sites that missed it are silent because their defaults happen to match".

**It also changes the recommended fix.** §5's step 2 (rename the six definitions) requires
editing other lanes' agents, and renaming a key that currently does nothing is a behaviour
change someone has to own. The shim above does not: three lines per action, entirely inside
this repo, no definition touched, and it makes the old name *work* rather than merely
*visible*. It is also the pattern already blessed in this codebase, so it needs no new
argument. **Prefer it** — it makes the bad state unrepresentable from the code side, whereas
renaming definitions leaves the next author free to write `check_domain` again.

Applying it to `run_discovery_checks` and `triage_detected_items` would make `check_domain`
and `target_domain` honoured — at which point the detector warnings this bug just switched on
should be **downgraded to `Deprecated`**, not deleted: `Deprecated` is exactly the field for
"recognised, still read, but the old name" (`ActionInputSpec.Deprecated`, and see
`BuildDeprecationMap`), and it keeps the audit honest instead of silent.

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

### 2c. `plan_sections.domain` — the third instance, and the only one on a live path

Found 2026-07-29 by intersecting the 151 undeclared actions with **what actually ran in the
last 24h**: 52 of them had traffic, and exactly one carried a key absent from its own source.

`page-build-handler`'s `plan_sections` step config is
`{domain, error_step, page_name, sections, site_id, work_item_id}`. `error_step` is a
framework key; every other key **is** in the spec — except `domain`. And:

- `grep -c domain platform/orchestration/actions/plan_sections_action.go` → **0**. The action
  never references the key at all.
- The spec's `Optional` contains **`pipeline`** — which **no live step sets**. The same
  half-landed rename, third site.
- The value is a dot-path, `"site_record.domain"`, so the *intent* was plainly to resolve a
  field. But `domain` is not in `Required ∪ Optional`, and **Strategy 0 iterates only those**
  (`action_inputs.go:359-360, 368`), so the path is never resolved. Fetched by nobody, used by
  nothing.

Opted in with `CheckConfig: true` rather than `ConfigKeys`, because this action reads nothing
from `StepConfig.Config` directly — everything arrives through `ExtractActionInputs` (`:620`),
so the spec is already a verified statement of what it reads and opting in asserts nothing new.
The step sets no `input_fields`, so the Strategy 1 escape hatch does not apply.

**This is now the route to the owed witness.** `plan_sections` ran **14 times in 24h** against
**0** for the other two, so the detector is exercised by traffic that is already happening,
without disturbing the deliberately quiesced improvement loop.

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

---

## Update 2026-08-08 — §2a's `[MEASURED] Nothing is mislabelled today` is now FALSE, and the framework gap behind all three instances is closed

**Fixed in code (`3f93456fd`), INERT until the next chassis roll.** Do not close it on the
commit — the bar is *fixed AND live*, and until the image ships the defect is still
reproducible. Council `Council-Submitted: 433de2c0-682f-4d8d-8c48-28637309f1ba`.
Workstream: `docs024_key_docs_latest/bugfix_136_config_key_aliases/`.

### 1. This file predicted its own trigger, and the trigger has fired

§2a said the containment was *"a coincidence of today's check-to-agent mapping"* and named
what would break it: *"The moment a check that propagates `dctx.Pipeline` is added to the
content or build agent's list … those findings are written as `design` and land in the wrong
queue."*

Since 2026-07-28, **two such checks have joined `completeness-discovery-agent`** —
`content_duplication` and `page_canonical_collision`, both of which set
`Pipeline: dctx.Pipeline` rather than hardcoding their own. That agent's config asks for
`check_domain: "content"`; the code read `check_pipeline`, found nothing, and took `"design"`.

```sql
SELECT item_type, pipeline, status, created_at::date FROM site_work_items
WHERE created_by='completeness-discovery-agent' AND pipeline='design';
--  page_canonical_collision | design | complete | 2026-08-04
--  page_canonical_collision | design | complete | 2026-08-04
--  capability_gap           | design | detected | 2026-08-04
--  capability_gap           | design | detected | 2026-08-03
```

> **CORRECTED 2026-08-08 — §2a's `[MEASURED] Nothing is mislabelled today` no longer holds.**
> Four rows, two still open. The measurement was honest when it was taken; what it measured
> was a check-to-agent mapping, and that is not a stable property. **The lesson is the one
> §2a itself half-stated: a figure that depends on which checks are registered where has a
> shelf life measured in days on this estate, and it should have been marked as such.**

The harm is not cosmetic. `countDispatchableWorkItems` (`work_items_common.go:198-211`)
filters `AND pipeline = $2` to answer *"is there unfinished promoted work on this site"*, so
a row under the wrong pipeline is invisible to the count that should see it.

### 2. Why two of three actions never got the shim — the framework gap this file did not name

§2 correctly identified `create_work_item`'s hand-rolled fallback as the pattern to prefer,
and asked why *"They wrote it on one action and not on the other two."* The answer is that
the framework offered no way to declare it.

`ActionInputSpec.Deprecated` **looks** like the field for this and is not. It is honoured in
exactly one place — `ExtractActionInputs` Strategy 3 — and there it does
`ExtractNestedField(collectedData, config[oldKey])`: it treats the old key's **value as a
dot-path into `collected_data`**. Correct for a *reference* alias (`site_id_field` →
`site_id`). For a *setting* — `check_domain: "content"` — it would look up a `collected_data`
key called `content`, find nothing, take the default, **and go quiet**, because
`UnknownConfigKeys` recognises `Deprecated` keys on purpose. Declaring it there is strictly
worse than not declaring it.

So the honest options were "hand-roll Go" or "do nothing", and two of three authors did
nothing. That is a framework defect wearing a per-action costume.

### 3. What shipped

`ActionInputSpec.DeprecatedConfigKeys` (old setting key → canonical key) + one shared reader,
`datahelpers.ResolveConfigSetting`, in the new `datahelpers/config_key_aliases.go`. Opt-in and
inert until a spec names it. Precedence is byte-for-byte the shim it replaces. Registered as
**SCR-006** and landmined in the same commit. Adopted by four actions:

| action | alias declared | value change |
|---|---|---|
| `run_discovery_checks` | `check_domain` → `check_pipeline` | **yes** — completeness's two propagating checks move `design` → `content`. design-agent asks for the value it already got; quality-agent runs six checks, none propagating |
| `triage_detected_items` | `target_domain` → `target_pipeline` | none — the sole caller asks for `"build"`, which is the default. **§2b's "3 agents" is now 1**: migration 286 removed the two child steps under RFC 006 |
| `create_work_item` | `item_domain` → `item_pipeline` | none — the hand-rolled shim converges onto the declared form |
| `execute_vision_prompt` | *(declares `ai_service`)* | none — it was a **false positive**: the action reads it via the shared `resolveAIServiceConfig` |

**Live audit, run against production `agent_definitions` after the change:**
UNKNOWN KEYS went **4 → 1**, and a new DEPRECATED section names all three renames.
The one left is `plan_sections: domain` — see §5.

### 4. How to verify, and where verification runs out

- **Reporting half needs no roll.** `./scripts/audit-config-keys.sh` runs `go run` from the
  tree against the live DB. UNKNOWN KEYS must be exactly `plan_sections: domain`; DEPRECATED
  must name `run_discovery_checks.check_domain` (3 steps), `triage_detected_items.target_domain`
  (1), `create_work_item.item_domain` (9).
- **Behaviour half, after the roll**, using a literal this change created (verified to appear
  nowhere else in the tree, so the grep discriminates):
  ```bash
  POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
  kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'config setting arrived via a deprecated alias'"   # 0 before, 1 after
  kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'ResolveConfigSetting'"                             # positive control
  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=-1 --since=24h | grep "deprecated alias"   # -l, NOT logs deploy/…
  ```
  `create_work_item` runs on live lanes today, so it exercises the **same shared helper** the
  two quiesced actions call — that is the available witness.
- **For `check_domain` / `target_domain` themselves there is NO live proof available**, and
  that is stated rather than worked around. The improvement loop is stopped by the owner
  ruling above, and *a verification that requires restarting something the owner stopped on
  purpose is not a verification worth having.* The claim chain is: shared helper proven live
  via `create_work_item` · per-action wiring proven by unit tests each killed by a deliberate
  mutation (see `config_key_alias_adopters_test.go`, including a sqlmock test that catches the
  action body being reverted while the declaration stays) · pod strings-grep proving the
  binary carries it.

### 5. Still owed on this file — deliberately not done, not forgotten

1. **The four mislabelled rows are NOT repaired.** Two `complete`, two `detected` (a queue
   with no consumer, `bugs_open/083`). Repairing them is a data edit on another lane's queue.
2. **`summary_template` (§4) is still biting** — two human-review items captioned with their
   own `item_type`. It is **not** an alias case: aliasing it to `summary` would ship a raw
   `{{.input_data.topic}}` string to the reviewer. It needs the render-or-literal decision.
3. **`plan_sections.domain`** stays in UNKNOWN KEYS. It is genuinely dead (the action reads
   neither `domain` nor the `pipeline` its spec declares), but `page-build-handler` is hot
   with several sessions on it, and the UNKNOWN line is the honest standing record until its
   owner deletes the key.
4. **`create_work_item`'s full opt-in** — blocked on adjudicating `summary_template`,
   `spec_fields` and `domain`. Note §3's read-list was wrong on one row: **`priority` IS
   read**, at `:144` via `datahelpers.GetIntField`, which a `config["…"]` grep cannot see.
5. **The definition renames.** Not needed for correctness now that the aliases are honoured;
   the DEPRECATED audit section is the standing migration list.

### 6. CORRECTION 2026-08-08 (same session, hours later) — I overstated the harm, and the enumeration I should have done first is below

> **§1 above says of the four mislabelled rows: *"The harm is not cosmetic"*, and cites
> `countDispatchableWorkItems` filtering `AND pipeline = $2`. That citation does not support
> the claim, and I made it in the council submission too (`433de2c0`).**

`countDispatchableWorkItems` has **exactly one caller** —
`triage_detect_items_action.go:189` — which passes `targetPipeline`, and the sole live
carrier asks for `"build"`. A row under `design` and a row under `content` are **equally
invisible** to that count. It cannot distinguish the two, so it cannot be evidence that
confusing them costs anything.

**The enumeration, done properly this time.** Every live consumer that filters
`site_work_items.pipeline` names a *specific* pipeline, and it is never `design` or `content`:

| consumer | predicate |
|---|---|
| `countDispatchableWorkItems` (1 caller) | `pipeline = $2`, always `"build"` today |
| `triageGatherFailures` | `pipeline <> 'diagnose'` — and excludes `capability_gap` by type anyway |
| `triageGatherCapabilityGaps` | **no pipeline filter at all** |
| `stale-work-item-reaper` | `status='triaged' AND pipeline='build'` |
| `report-dispatch-loop` claim/reap | `pipeline='reports'` |
| `diagnose-dispatch-loop` claim/reap | `pipeline='diagnose'` |

Live distribution: `build` 3475 · `content` 321 · `design` 118 · `diagnose` 30 ·
`maintenance` 1. **Nothing dispatches, reaps or counts `design` or `content` distinctly.**

**So the accurate statement, which is weaker than §1's and is the one to quote:**

- **The four rows ARE mislabelled.** That is measured and is not in doubt: an agent
  configured `content` filed them `design`.
- **No live consumer distinguishes those two values today**, so there is no demonstrated
  downstream failure — the label is wrong, and nothing currently reads it in a way that
  makes the wrongness cost anything.
- **The exposure is therefore the same shape as the original bug, one level out:** the
  labels are correct-or-not by luck, and the first consumer to filter on `design` vs
  `content` inherits whatever the default happened to write. That is a real reason to fix
  it and a poor reason to call it urgent.

**Nothing about the fix changes.** The config was inert (measured), the framework gap was
real (read from the code), and honouring what the config says is right whether or not
today's consumers can tell. What changes is the *argument*, and an argument that overstates
its evidence is the thing this estate is least willing to carry forward — the more so
because §2a's original `[MEASURED]` claim was over-trusted in exactly the same way, by me,
five hours earlier in this same file.

**The check that would have caught it at t=0, and it is one grep:** before citing a
consumer as harmed, `grep -rn "<function>" --include=*.go` for its **callers** and read
what each passes. I read the predicate (`AND pipeline = $2`) and inferred the harm from the
*shape* of the query without asking what `$2` is ever bound to. A parameterised filter
tells you what a query *can* discriminate, never what it *does*. Logged in `WRONG_CALLS.md`.

### 7. Council round 1 — APPROVED, and four of the objections were worth acting on

`433de2c0-682f-4d8d-8c48-28637309f1ba` · **approved with 2 advisory objections, none
high-severity** · 12 seats reviewed, 5 abstained, `unreadable: 0`, ~8 minutes.

The `architecture` seat explicitly confirmed the scope call: the change *"satisfies the
2026-07-29 owner-ruling condition for treating additive/inert/opt-in struct fields as a
normal gate"* — LANDMINE and register entry in the same commit, plus the parity test — and
called it *"the fix taking the shared-mechanism route deliberately, correctly, and with the
guardrails an RFC would otherwise force"*. So this did **not** need an RFC, and the reason
it did not is the same-commit registration, not the smallness of the diff.

**Acted on, with the result:**

1. **`guardian` (medium) — "confirm the pipeline-consumer search is complete, not just a Go
   grep; a definition-level consumer (an `agent_definitions` `query_database` step filtering
   `pipeline='design'`) would not show up."** Correct, and it is the same gap §6 above found
   independently. Done: every live step-config query touching `site_work_items` and
   `pipeline` names a *specific* pipeline — `build`, `reports`, `diagnose` — and **none names
   `design` or `content`**. Combined with §6's Go-side enumeration, the consumer search is now
   complete across both surfaces.
2. **`debug_historian` (medium) — the pod-grep proof is positive-only, validated against
   source uniqueness rather than the running binary, and `-l app=agent-chassis` may cover a
   fraction of the pods running that image.** Both true, and the second is worse than the
   seat guessed: **`-l app=agent-chassis` returns 2 pods; 25 running pods carry an
   agent-chassis image (34 including non-Running).** The label selector covers **8%** of the
   surface. §4's recipe is superseded by §8 below, which enumerates by IMAGE and carries both
   controls.
3. **`reuse_agent` (low) — "confirm no other ad-hoc config-rename shim exists platform-wide
   that should have converged here rather than spawning a third pattern."** Searched. Two
   candidates, one real: `section_editor_actions.go:56-59` uses the existing `Deprecated`
   field **correctly** (both entries are reference aliases), so it is not a candidate. But
   `resolveAgentTypeForSpawn` (`spawn_actions.go:3154-3163`) **is a fourth hand-rolled
   instance** — `group_type` → `agent_type` is a literal-setting alias, exactly this class,
   hand-rolled and invisible to the audit. **Not converged, and measured before deciding
   that:** `group_type` and `group_type_field` are set by **zero** live steps
   (`agent_type_field` by 3), so it guards a config shape nobody writes. Recorded as a
   convergence candidate with no live exposure rather than pulled into this change.
4. **`prior_art_librarian` — "not independently confirmed there is no second consumer of
   `Deprecated`."** Confirmed: `spec.Deprecated` is *honoured* in exactly one place
   (`ExtractActionInputs` Strategy 3, `:521`) and *reported* in four more
   (`UnknownConfigKeys:226`, `ListDeclaredConfigKeys:344`, `GenerateInputContract:743`,
   `cmd/config-key-audit:244`). `registry.go`'s `def.Deprecated` is a **different field** — a
   bool on the action registry entry, not the spec's map — and is unrelated.

**Recorded, not acted on:**

- **`editquality` (medium ×2) — the parity test, the behaviour-preserving test, the LANDMINE
  and the concept-register entry are claimed in the rationale but absent from the `edits`
  list, so "the mitigations described in prose do not exist in the plan".** The seat is right
  about the *submission* and wrong about the *world*: all four exist and shipped in
  `3f93456fd`. The cause is the schema's 8-edit cap — I spent all eight on code and described
  the rest in prose. **The lesson is that a reviewer can only see the edits list, so a
  mitigation named only in the rationale reads as fiction.** Next time: spend an edit slot on
  the test file, or say plainly "shipped in the same commit, not listed, cap reached".
- **`editquality` (low) — edit 8 (`ai_service`) is scope creep, not on the causal path.**
  Fair on minimality. I keep it: the false positive was in the very report this change is
  judged by, and a report with a known-false line is one people stop reading.
- **`architecture` (low) — two fields named `Deprecated` and `DeprecatedConfigKeys` with
  opposite resolution semantics is "a durable readability tax on every future action
  author".** Agreed, and unmitigable beyond what shipped (the LANDMINE, the parity test, and
  the doc comment on each field pointing at the other). Noted as an accepted cost.

### 8. The verification recipe, corrected (supersedes §4's pod-grep)

Enumerate by **image**, not by label, and carry both controls. Measured pre-roll on
2026-08-08 — `new=0` (not yet shipped), `pos_control=1` (the grep works and this is the right
binary), `neg_control=0` (it is not matching everything):

```bash
PODS=$(kubectl -n ai-persona-system get pods -o json | python3 -c "
import json,sys
d=json.load(sys.stdin)
print(' '.join(p['metadata']['name'] for p in d['items']
      if p.get('status',{}).get('phase')=='Running'
      and any('agent-chassis' in c.get('image','') for c in p['spec'].get('containers',[]))))")
echo \"pods carrying the image: $(echo $PODS | wc -w)\"   # 25 on 2026-08-08; -l app=agent-chassis sees 2
for P in $PODS; do
  kubectl -n ai-persona-system exec "$P" -- sh -c "
    echo -n '$P new=';        strings /app/agent-chassis | grep -c 'config setting arrived via a deprecated alias'
    echo -n '$P pos_control='; strings /app/agent-chassis | grep -c 'Using deprecated config pattern'
    echo -n '$P neg_control='; strings /app/agent-chassis | grep -c 'zzz_invented_control_string_136'"
done
```

`new` must go 0 → 1 on every pod. `pos_control` (Strategy 3's long-live warn) must be 1
**before and after** — it proves the grep and the path to the binary, which a
new-string-only check cannot. `neg_control` must be 0 always.

### 9. 2026-08-08 (evening) — LIVE on v1.0.1267. Binary proven; runtime behaviour still UNWITNESSED, and the witness I designed could never have worked

**The code is live.** Verified with §8's corrected recipe — enumerate by image, both controls
— and it is a genuine before/after across the roll, same command both times:

| when | `new` | `pos_control` | `neg_control` |
|---|---|---|---|
| pre-roll (v1.0.1264/1266, banked in §8) | **0** | 1 | 0 |
| post-roll, `agent-chassis-88f79d88c-4v2d2` (v1.0.1267) | **1** | 1 | 0 |
| post-roll, `agent-build-dispatch-loop-e4e60240-xdg5w` (v1.0.1267) | **1** | 1 | 0 |

`ResolveConfigSetting` resolves 2 in both. **Roll coverage at 17:05Z: 17 of 19 pods carrying
a chassis image are on v1.0.1267**; one lags on v1.0.1266 and one on v1.0.1264, so the fleet
is not uniformly on it yet. My commit is `2026-08-08 15:41:13 UTC` — stated in UTC because
this machine is BST and comparing a BST `git log` against a UTC `kubectl` is how a live fix
reads as un-shipped.

> **CORRECTED 2026-08-08 — §4 and the lane PLAN both said `create_work_item` running on live
> lanes was "the cheapest honest live proof" and "the available witness". THAT IS WRONG, and
> wrong in this bug's own signature way.**
>
> **All nine live `item_domain` carriers set `"build"` — which is exactly
> `create_work_item`'s hardcoded default.**
>
> ```sql
> SELECT ad.type, e.v->'config'->>'item_domain' FROM agent_definitions ad,
>        jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
>  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
>    AND e.v->>'action'='create_work_item' AND e.v->'config' ? 'item_domain';
> --  all 9 rows: build
> ```
>
> So a `create_work_item` run produces `pipeline='build'` **whether the alias was honoured or
> the default was taken**. The observation cannot come out otherwise, and a measurement that
> cannot come out otherwise is not evidence — the rule this estate already wrote down after
> two dated-and-marked figures turned out to be unfalsifiable. I proposed it as a witness
> without asking what the disconfirming result would have looked like.

**The only discriminator is the log line**, which fires solely when an alias supplies the
value. Swept all 23 pods carrying the image, `--since=3h`: **0 hits — and `create_work_item`
itself also appears 0 times in every one of those logs.** The positive control is zero, so
the sweep says *"these executions are not visible in these logs"*, **not** *"the alias did
not fire"*. One `create_work_item` run did happen on the new binary (`tool-improver`,
orchestration `2801f868`, row at 16:39:36Z, `pipeline='build'`) and its execution appears in
no pod log I could reach.

**Honest state, therefore:**

| claim | status |
|---|---|
| the binary carries the change | **PROVEN** — pod-grep, both controls, before/after a roll |
| the declarations join real live definitions | **PROVEN** — audit UNKNOWN KEYS 4 → 1 against production `agent_definitions` |
| the alias is honoured at runtime | **UNWITNESSED** — and not obtainable from `create_work_item`, whose every live carrier's value equals the default |
| `run_discovery_checks` / `triage_detected_items` behaviour | **UNWITNESSED** — quiesced by owner ruling; unchanged from §4 |

**What would actually discriminate**, for whoever picks this up:
1. **Find where `create_work_item` logs.** The action does log; those lines are not in the
   pod logs I swept, so the sink or the level is the question. Answer that and the warn line
   becomes reachable — it fires on *every* call from those nine agents.
2. **Or change one carrier's `item_domain` to a non-default value** (`"content"`) and watch
   the next row's `pipeline`. That is a definition edit on another lane's agent and I have
   not done it — but it is the only single-step live discriminator that exists, and it is
   cheap and reversible.

**This bug stays OPEN**, and now for a sharper reason than "not yet rolled": it is rolled,
and the runtime half is unproven. Per the owner ruling of 2026-08-06 a finished bug stays in
`bugs_open/` regardless.

### 10. 2026-08-08 (late) — §4's biting key RESOLVED by owner decision: A + D, template rendering declined

**Owner decision 2026-08-08**, options presented with the re-measured facts: **A** (static
`summary` literal, drop `summary_template`) **+ D** (recaption the two existing rows).
**Option C — implementing template rendering in `create_work_item` — was explicitly
declined**: all ten other live `create_work_item` steps use a static summary, the sole
would-be consumer has executed zero orchestrations, and a template naming a missing key
fails on the path to a human reviewer. If interpolated captions ever have real demand,
build it then, on evidence.

**Applied and verified live** (migration `343_grounded_explainer_summary_literal.sql`;
DB config, effective immediately, no roll):

- definition: `summary_template` gone, `summary` literal in — confirmed by re-reading the
  live row;
- both `grounded_draft_review` items now read *"Grounded explainer draft ready for review"*
  — the data UPDATE was keyed on `summary = item_type`, so a row a human had since retitled
  would have been left alone;
- the repo seed (`224_grounded_explainer_agent.sql:183`) corrected in the same commit, so a
  replay cannot reintroduce the dead key — the `bugs_open/134` lesson.

**The topic is NOT in the repaired captions, and cannot be:** the same step's `spec_fields`
is also a dead key (§3), so both rows' `spec` is completely empty — the topic was never
captured anywhere. The generic caption is the truth of what those rows can say.

**§5's still-owed list, updated:** item 2 (`summary_template`) is DONE. Items 1 (the four
mislabelled pipeline rows), 3 (`plan_sections.domain`), 4 (`create_work_item` full opt-in —
note `spec_fields` remains dead and remains listed there), and 5 (definition renames /
`resolveAgentTypeForSpawn` convergence candidate) stand. The runtime witness question from
§9 also stands.
