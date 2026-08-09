# 236 — `hero_deployed` / `logo_deployed` carry no `image_url` by the time the readers run, so hero and logo URLs are never set

**Filed 2026-08-09** by the `rfc012_await_findings` lane, at the owner's direction to clear the
lane's residual problems. **Status: OPEN, root cause NOT established — see §5, which is the
honest part of this file.**

## 1. What is established, first-hand, from live rows

The 08-06 handoff (§2) recorded this as a *hypothesis that could not be tested*, because
`hero_deployed` appeared in **0 of 1,667** retained `orchestration_states` rows and the table keeps
~24h. **That is no longer true: the keys are present in live rows today**, and the shape is the
one the hypothesis feared.

```sql
SELECT count(*) FILTER (WHERE collected_data ? 'hero_deployed') AS hero,
       count(*) FILTER (WHERE collected_data ? 'logo_deployed') AS logo,
       count(*) AS total_retained
FROM orchestration_states;
--  hero | logo | total_retained
--     2 |    2 |           4757     (2026-08-09 ~14:35Z)
```

**The decisive row** is `3e46be5b-8788-447b-9643-e32ae33f601b` (`site-work-orchestrator`,
`cookly.uk`), because its logo request **did** come back — so this is not a story about a timeout:

```json
// collected_data->'logo_deployed'
{
    "response": {
        "data": {
            "files": ["/assets/images/logo.png"],
            "domain": "cookly.uk",
            "success": true,
            "file_path": "/assets/images/logo.png",
            "commit_message": "Deploy logo image for cookly.uk",
            "timestamp": "2026-08-09T13:54:36Z"
        }
    },
    "response_status": "complete",
    "response_received_at": "2026-08-09T13:54:37Z"
}
```

**No `image_url`. No `output_path`. No `size_bytes`** — none of the three keys
`DeployImageAssetAction` assigns onto its own result at
`platform/orchestration/actions/deploy_image_asset_action.go:369-371`:

```go
result, err := sendGitCommitRequest(ctx, params, domain, filesMap, purpose, logger)
if err != nil { return nil, err }
result["image_url"]   = processed.Paths.RelativeURL
result["output_path"] = processed.Paths.FilePath
result["size_bytes"]  = len(processed.Data)
```

And the readers need exactly the key that is missing —
`platform/orchestration/actions/v3_site_actions.go:1020` and `:1031`:

```go
if heroDeployed, ok := params.CollectedData["hero_deployed"].(map[string]interface{}); ok {
    if imageURL, ok := heroDeployed["image_url"].(string); ok && imageURL != "" {
        renderCtx.ContentData["hero_url"] = imageURL
        ...
    }
}
```

Both guards are `ok`-checked with **no else branch**, so a miss is silent — no log, no error, no
work item. `assemble_from_library.go:452` is the third site of the same shape.

**The consequence is visible on the same row:** `collected_data` carries neither `hero_url` nor
`logo_url`, so the later "fallback" read of those plain keys (the one the 090 run found and the
code labels `(fallback)`) has nothing to find either.

```sql
SELECT collected_data ? 'hero_url' AS has_hero_url,
       collected_data ? 'logo_url' AS has_logo_url
FROM orchestration_states WHERE orchestration_id='3e46be5b-8788-447b-9643-e32ae33f601b';
--  f | f
```

The other two rows (`22fb157a-…`, and the hero half of `3e46be5b-…`) show
`{"response_status":"complete","response_received_at":"…"}` with **no `response` key at all**, and
both runs carry `error = "Request <id> timed out after 3 retries"`. Those are a *different* state
and are **not** the evidence for this bug — noted so nobody quotes them as such.

## 2. What the readers being silent costs

Whatever the cause, the rendered page gets no hero and no logo, and nothing anywhere records that
it happened. This is the `bugs_open/012`-family shape: the status says `complete`, the artefact is
missing content. **Trust the rendered artefact, not the status.**

## 3. Which sites and how often — NOT MEASURED

Two runs in the retention window is all the evidence there is. `[UNMEASURED]` whether this is
every image deploy, or these two. The retention wall that blocked the 08-06 investigation still
blocks the historical question; only forward observation can size it. Do **not** quote "2" as an
incidence rate — it is the count of rows that happened to be in the window when someone looked.

## 4. Fix candidates — READ §5 FIRST, none of these is licensed yet

Ordered by what makes the bad state unrepresentable rather than by effort:

1. **Make the readers loud.** An `else` on each of the three guards, naming the key it wanted and
   what the map actually held. This is contained, is in nobody's decision territory, and would have
   turned a silent five-week defect into a log line on day one. **Weakest fix, strongest
   cost/benefit** — it fixes the *invisibility*, not the defect.
2. **Have the readers accept the merged shape** — read `image_url`, then
   `response.data.file_path` / `response.data.files[0]`. Cheap, and wrong as a first move: it
   encodes the merged shape at three call sites, which is precisely the "patch one call site by
   hand" pattern the RFC_012 census already found evidence of at `unified_extractor.go:200`.
3. **Fix whatever removes the action's own keys** — the real fix, and it cannot be written until
   §5 is answered.

## 5. THE ROOT CAUSE IS NOT ESTABLISHED, and the obvious theory is REFUTED by the code

The natural theory — *"the awaited-response merge overwrites the key and discards the action's own
workings"* — **is not supported by the merge code**. `coordinator.go:2719-2748` reads the existing
map and adds to it rather than replacing it, on both the step-name and the `output_field` branch:

```go
existingData, exists := state.CollectedData[stepName].(map[string]interface{})
if !exists { existingData = make(map[string]interface{}) }
existingData["response"] = normalisedData
existingData["response_received_at"] = ...
existingData["response_status"] = "complete"
state.CollectedData[stepName] = existingData
// …and the same preserve-then-add shape again for outputField
```

And the action's result *is* written under the output field before the await:
`storeActionResult` (`coordinator.go:1863`, called at `:1795`) sets
`state.CollectedData[step.OutputField] = result`, and it runs **before** `processAwaitResponse`
(`:1839`), which then persists the state.

**So on a straight reading, `image_url` should still be there — and in the live row it is not.**
That gap is the bug, and I am deliberately not guessing at it. Candidate directions, all
`[UNVERIFIED]`, offered as starting points and not as findings:

- the response handler may not be reading the same state instance/version that
  `storeActionResult` wrote (an optimistic-lock reload, or a persist that did not include
  `CollectedData`);
- `stepName` at response time may not equal `state.CurrentStep` at store time, so the preserve
  branch finds an absent or different map;
- a normalisation step between the two may rebuild the map rather than merge into it;
- the `isAgentResponse` branch may not be the branch these rows took at all — `:2666` and `:2743`
  write the same two stamps from other paths, and I have not established which one produced these
  rows.

**Per CLAUDE.md this is exactly the "cause is not where the symptom is" case**, so it goes to the
diagnosis loop rather than into a fix. A `090` run on the earlier framing came back
`UNVERIFIABLE` (`dce40cf4-5a8a-4316-93c0-0f3c37d2f3a7`) — correctly, because at the time *"the
evidence would not grow"*. **The evidence now exists**, which is the one thing that changed and
the reason a re-run is worth its credits rather than a repeat.

## 5a. The 090 run: UNVERIFIABLE again, and this time it is the HARNESS, not the question

`RUN_CORRELATION_ID=074beb8a-adb4-4074-905a-cb0f857e7f85`, filed with §5's framing (which names
the refutation, so the loop was not asked to confirm a story). Verdict: **`UNVERIFIABLE`,
`stopped: scope-not-narrowing`, `is_fix: false`** — *"Hand to a human with the full trail; do NOT
auto-conclude."* **Read it correctly: it did not refute anything.** It says, in its own words, what
it still needed, and both items are tooling failures rather than absent evidence:

1. **It could not read the functions.** *"the actual bodies of `storeActionResult`,
   `processAwaitResponse`, and `applyResponseToState` (now located at L2635-2764 via code_request,
   but not yet read in full — only the single line `isAgentResponse := false` is visible)"*. The
   symbol *"failed to load from coordinator.go in this checkout"*. So the loop was reasoning about
   a merge it could not open — which is, with some irony, **the exact mistake this bug's own
   `WRONG_CALLS` entry records me making by hand**.
2. **Its SQL could not address the table.** *"The bundle's own data_request for this row failed
   with `column "id" does not exist (SQLSTATE 42703)` — the `orchestration_states` table isn't in
   the bundle's Schema section, so its real primary-key/id column is unknown and must be confirmed
   by a human before requerying; guessing again would likely fail the same way."* It queried
   `WHERE id = '3e46be5b-…'`; the column is **`orchestration_id`**, which the symptom text gave it.

**Finding for whoever owns the diagnosis loop, and it is bigger than this bug.** The bundle's
"Schema (live tables)" section exists precisely so *"the verdict writes a correct data_request
instead of guessing a relation that does not exist"*
(`diagnose_assemble_bundle_action.go:303-309`, fed from `runtime.schema`). **`orchestration_states`
— the platform's central run-state table — is not in it**, so any hypothesis whose evidence lives
there is unfalsifiable by the loop for a reason that has nothing to do with the hypothesis. That is
a standing blind spot, not a one-off. `[UNVERIFIED]` where the table list is built: I did not
establish what populates `runtime.schema`, and deliberately do not guess — the loop's own error
message is the evidence here, not my reading.

**Both 090 attempts on this symptom are now UNVERIFIABLE for different reasons** — the first
(`dce40cf4`) because the evidence did not exist, this one because the harness could not reach
evidence that does. Neither is a refutation, and nobody should quote the pair as one.

## 6. Why this bug is filed and not fixed

The only fix that addresses the cause may land in the awaited-merge design — RFC_012's `(a)`/`(a′)`
question, which is **an open owner decision**, not a thread's call. Candidate 1 (make the readers
loud) is contained and is the sane thing to ship regardless of how the decision goes; it is
deliberately left for whoever picks this up with a mandate, so that this file records evidence
rather than smuggling a design choice.

## 7. How to verify, when someone fixes it

Forward observation only; the historical rows cannot answer it.

```sql
-- the shape, caught while it is in the window
SELECT orchestration_id, owner_agent_type,
       collected_data->'hero_deployed' ? 'image_url' AS hero_has_url,
       collected_data->'logo_deployed' ? 'image_url' AS logo_has_url,
       collected_data ? 'hero_url' AS hero_url_set
FROM orchestration_states
WHERE collected_data ? 'hero_deployed' OR collected_data ? 'logo_deployed'
ORDER BY created_at DESC;
```

Then the artefact, which is the claim that matters: the rendered page carries a hero image and a
logo. **A green run is not evidence** — that is what this bug looked like for five weeks.

## 8. Sources

- Live rows `3e46be5b-8788-447b-9643-e32ae33f601b` and `22fb157a-322b-4c56-bc00-51e549db3060`,
  read 2026-08-09 ~14:35Z. Both expire from retention within ~24h of that time.
- `platform/orchestration/actions/deploy_image_asset_action.go:362-371` (the writer);
  `platform/orchestration/actions/v3_site_actions.go:1020`, `:1031`,
  `platform/orchestration/actions/assemble_from_library.go:452` (the three silent readers);
  `platform/orchestration/coordinator.go:1795`, `:1863` (`storeActionResult`), `:1839`
  (`processAwaitResponse`), `:2719-2748` (the preserve-then-add merge that refutes the obvious
  theory).
- Prior history: `rfc012_await_findings/HANDOFF_2026-08-06_continue_here.md` §2 (the hypothesis and
  the 090 verdict), `architecture_review/CENSUS_2026-08-07_rfc012_await_step_readers.md` (the Go
  side's "3 breaks, silent").
