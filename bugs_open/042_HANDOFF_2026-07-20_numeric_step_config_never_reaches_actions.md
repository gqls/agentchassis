# 042 — numeric step config never reaches actions; string literals are misread as references

**Filed 2026-07-20** (vetcomparison thread). **Status: OPEN.** Fleet-wide.
Diagnosis-loop correlation `f155b0c4-881b-4369-abe4-569d7b2ad4c8` (filed; the loop has been
failing to reach a verdict today — see §Caveat — so this case file is the durable record).

## One-line

`ExtractActionInputs` only ever reads step config through `config[field].(string)`, and even
then treats the string as a **reference to resolve against `collectedData`** — never as a
literal. So a numeric config value is silently dropped and the action runs on its Go fallback.

## Why nobody noticed

**The seeded value equalled the code default.** `render_news_json` config carried
`max_age_hours: 72`; `RenderNewsSectionAction` calls `inputs.GetInt("max_age_hours", 72)`.
Config and behaviour agreed, so the config looked wired up and load-bearing when it was
decorative. It took *changing* the value to discover that changing the value does nothing.

> **Transferable lesson (also added to 016b §9): a config setting that matches its code default
> proves nothing about whether it is wired up.** To test whether config is live you must set it
> to a value that would produce visibly different behaviour, then observe the behaviour — not
> re-read the config.

## Mechanism

`platform/orchestration/datahelpers/action_inputs.go`, `ExtractActionInputs`. Every branch that
consults step config:

| line | branch | reads config as |
|---|---|---|
| :126 | Strategy 0 — explicit dot-paths | `config[field].(string)` |
| :144 | `input_fields` | `config["input_fields"].([]interface{})` |
| :180 | deprecated `*_field` | `config[oldKey].(string)` |
| :233 | Strategy 4 — remaining refs | `config[field].(string)` |

There is **no branch that takes a literal config value**. Two consequences:

1. **Numbers and booleans** fail the type assertion and never enter `result.Values`. The action
   receives whatever fallback its call site passes (`GetInt("max_age_hours", 72)`).
2. **Plain strings without a dot** are treated as single-segment *references* (:253) and looked
   up as keys in `collectedData`. A literal like `"veterinary"` or `"vetcomparison.uk"` resolves
   to nothing unless a collectedData key happens to share that name.

## Evidence — predicted, then observed

Set the live `content-feed-orchestrator` row's `render_news_json.config.max_age_hours` to `720`
(30 days). The site's two feed items are ~483 h old, so 720 must include them and 72 must not.

- The run carried the new value — from its own `initial_request_data.agent_config`:
  `{"site_id":"input_data.site_id","max_items":6,"page_name":"index","max_age_hours":720}`
- The renderer's own query, run verbatim against the live DB with 720, **returns both items**
  (status `relevant`, relevance 55).
- The run nevertheless rendered `item_count: 0`, `items: []` — **exactly what a 72 h window
  predicts**, and `check_has_news` then routed `0 → complete`, skipping `commit_news`, so
  `data/latest-news.json` was never published at all.
- Deployed binary confirmed to contain the current query (per CLAUDE.md, checked against the
  pod, not git): `strings /app/agent-chassis | grep -c loadNewsItems` → 5.

`max_items: 6` in that same config has **never** been read either — it merely equals its fallback.

## Blast radius

Any action tuned by a numeric step config value is silently running on its Go default, and the
config reads as though it works. This is worse than a value being wrong: the DB is the documented
tuning surface, so operators change a number, observe nothing, and conclude the number was
already right. Every `GetInt`/`GetBool` call site with a non-trivial default is a candidate.

`docs/agent_docs/sql_for_agents/090_content_feed_orchestrator.sql` now carries `720` with a loud
inline comment recording that it is inert until this is fixed — the intent survives and takes
effect on the fix, without being mistaken for working.

## Fix candidates

1. **Add a literal fallback to `ExtractActionInputs`** (recommended): after the reference
   strategies, for any `allFields` entry still unset, take `config[field]` as a literal value
   regardless of type. Preserves existing reference semantics — references already win, because
   they run first and the literal pass only fills what is still missing.
2. **Type-aware handling**: treat non-string config values as literals immediately; keep strings
   as references. Narrower, but leaves literal *strings* still broken (see the sibling exporter
   case, correlation `55dc0fa4-116c-40d6-90b2-bfad9ad73692`, where a literal domain string does
   not reach the action).
3. **Explicit literal marker** in config (e.g. `{"$literal": 720}`). Most precise, most invasive
   — every existing config would need auditing.

Option 1 is the smallest change that makes the documented tuning surface actually work.

## How to verify a fix

Do **not** re-read the config. Set a numeric config value to something whose effect is visible,
run the action, and check the *behaviour*:

```
-- set render_news_json.config.max_age_hours = 720 for content-feed-orchestrator, then after a run:
SELECT collected_data->'news_render_result'->'item_count'
FROM orchestration_states WHERE ... ;      -- must be > 0 while items older than 72h exist
```

Then confirm the artefact, not the status: `curl -s https://<site>/data/latest-news.json` must
return 200 with a non-empty `items` array.

## Related

- Sibling, same family (literal not reaching an action), filed separately:
  `DirectoryExportAction` aborts on an empty domain although `scheduled_tasks.input_data` carries
  one — correlation `55dc0fa4-116c-40d6-90b2-bfad9ad73692`. **Consequence: vetcomparison.uk's
  practice directory has stopped refreshing since 2026-07-17** while the site still serves the
  last good file, so nothing looks wrong.
- `/bugs_open/027` — its server-render fix is committed (`1005e1af2`) but **not deployed**
  (`persistNewsSectionHTML` absent from the running binary), so 027 stays open on the
  fixed-AND-live bar.

## Caveat on the diagnosis-loop trail

Three of four diagnosis runs filed from this thread today died without a verdict —
`reaper: stale AWAITING_RESPONSES for >90 min` (queued ~32 min behind a busy cluster),
`reaper: stale EXECUTING_STEP for >4h; step=route`, and one failure at `call_diagnoser`. Bundles
were produced in each case; no verdict was. **A filed diagnosis is not a delivered one** — check
for a terminal verdict before treating a correlation id as an answer, and note the intake record
still reads `awaiting_diagnosis` after its orchestration has died, which makes a refile look like
a duplicate. Close the stale intake first, then refile.
