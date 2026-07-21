# 042 — numeric step config never reaches actions; string literals are misread as references

> **CLOSED 2026-07-21 — FIXED AND LIVE in v1.0.1144.** Verified against the pod
> (`strings /app/agent-chassis | grep -c 'took literal scalar config value'` -> 1, the
> Strategy 5 marker) AND behaviourally: with the fix live, `content-feed-orchestrator`'s
> `max_age_hours: 720` finally reaches `RenderNewsSectionAction`, so vetcomparison.uk's
> `/data/latest-news.json` now publishes 3 CMA items that are 460h+ old — impossible under the
> silent 72h fallback this bug describes. The numeric config now demonstrably drives behaviour.
> Fix commit `4ac86e345` (Strategy 5, non-string scalars only) + tests. Council review was
> attempted but the submission was rejected at intake for including a file-`create` edit (see
> WRONG_CALLS 2026-07-20); the fix is covered by its own unit tests instead.
> **The literal-STRING half is NOT fixed** — an unresolved literal string still never reaches an
> action (the exporter's domain, `55dc0fa4`). That is deliberately out of scope here; if it needs
> a home, open a new case rather than reopening this one.


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

> **CORRECTED 2026-07-20, and this correction is the point.** Option 1 as first written above
> ("take `config[field]` as a literal **regardless of type**") carries a regression risk I had
> not thought through: a genuine reference that *fails to resolve* — say `"current_item.id"` on
> an iteration where that data is absent — would stop being absent and silently become the
> literal string `"current_item.id"`. A wiring bug would turn into a plausible-looking value.
> That is a worse failure than the one being fixed, because it is invisible.
>
> **Implemented instead: non-string scalars only.** References are always strings, so restricting
> the literal pass to `bool`, the int/uint families, `float32/64` and `json.Number` cannot alter
> how any existing reference resolves — it only fills fields currently dropped on the floor. An
> unresolvable string is still left ABSENT, on purpose, so broken references stay visible.
> Composite values (objects, arrays) are excluded too: no evidence they were ever meant as
> literals here.
>
> This leaves the literal-*string* case unfixed, which is a real gap — the sibling exporter bug
> (`55dc0fa4`) is exactly that shape, where a literal domain string never reaches the action. It
> needs its own change and its own argument, not a free ride on this one.

## Status

**Fix implemented 2026-07-20** in `action_inputs.go` as Strategy 5 (non-string scalars only),
with regression cover in `action_inputs_literal_test.go`. Five tests pass; `datahelpers` and
`actions` packages both build and pass. Submitted to the council gate as
`712be028-1c57-4e90-a0b0-09eb9742fc9a`.

**Go change — inert until the next image roll.** Until then the effective news window remains
72 h and `data/latest-news.json` is still not published for vetcomparison.uk. Verify after the
roll against the *pod*, not the tag, then against the *artefact*, not the status.

The tests deliberately configure values that DIFFER from their fallbacks. A test that reuses the
default would pass whether or not the plumbing works — which is precisely how this defect
survived.

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

- Sibling, ~~same family (literal not reaching an action)~~, filed separately:
  `DirectoryExportAction` aborts on an empty domain although `scheduled_tasks.input_data` carries
  one — correlation `55dc0fa4-116c-40d6-90b2-bfad9ad73692`. **Consequence: vetcomparison.uk's
  practice directory has stopped refreshing since 2026-07-17** while the site still serves the
  last good file, so nothing looks wrong.
  > **CORRECTED 2026-07-21 — this is NOT the same family, and the grouping was never checked
  > against the code.** `DirectoryExportJSONAction` does not use `ExtractActionInputs`; it reads
  > `config["domain"].(string)` directly. The domain is an ordinary string — the problem is that it
  > is one nesting level too deep, because the scheduled task's `input_data` was authored as a full
  > message envelope and the scheduler's `fireTrigger` wraps it a second time
  > (`input_data.input_data.domain`). Fixing the literal-string half of `ExtractActionInputs` would
  > do nothing here. **Full root cause, fix and behavioural verification: `bugs_open/054`
  > (FIXED & LIVE 2026-07-21, data + seed, no image roll).** Caught by reading the failing action
  > and tracing the live `collected_data` of run `6271b72d`. The literal-*string* half of
  > `ExtractActionInputs` (this file's §Fix candidates) remains a genuine, separate gap.
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
