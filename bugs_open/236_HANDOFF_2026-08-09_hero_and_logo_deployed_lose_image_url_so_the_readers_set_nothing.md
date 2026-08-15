# 236 — `hero_deployed` / `logo_deployed` carry no `image_url` by the time the readers run, so hero and logo URLs are never set

**Filed 2026-08-09** by the `rfc012_await_findings` lane, at the owner's direction to clear the
lane's residual problems. **Status: OPEN — ~~root cause NOT established~~ the MECHANISM is
CONFIRMED as of 2026-08-14** (the park discards in-memory `CollectedData`; see the final
contribution) — **the fix is an RFC_012 owner decision, not a patch for this file.** §5 remains
the honest record of how the cause was narrowed.

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

---

## 5b. 2026-08-11 — the HARNESS is fixed and proven; the EVIDENCE has now expired

Contributed by the `diagnosis_schema_visibility` lane (commission item 5). **Two separate
results here — read them apart, because one is good news and one changes what this bug needs.**

### The harness blocker named in §5a is FIXED and verified on this bug

Commission item 5 shipped (`5f8a326fc`, council-approved `df9dae6c`, live on
`agent-chassis:v1.0.1284`, pod-verified with a negative control). The diagnosis bundle's Schema
section was passing through a relevance include (`site%|page%|content%|flow%`) that selected
**26 of 433** live tables — and **five of the six tables the gather itself renders rows from**
fell outside it, `orchestration_states` among them. The section also never said it was filtered,
so a filtered-out table and a non-existent table read identically. That is the whole of why
`074beb8a` said the id column "must be confirmed by a human".

Re-run **`90f6f55f-c014-4537-880c-0f1ae2b82e0b`** (2026-08-11), all five bundle iterations:

| check | result |
|---|---|
| any `42703` in the bundle | **none, all 5 iterations** |
| `orchestration_states(…)` columns in the Schema section | **present, all 5** |
| the "listing is FILTERED" notice | **present, all 5** |
| a `data_request` against `orchestration_states` | **EXECUTED** — returned `(0 rows)`, correct columns (`orchestration_id`, `owner_agent_type`), no error |

**So §5a's item 2 is closed.** (§5a's item 1 — the loop could not read the coordinator function
bodies — is also clear: the index is fresh and carries all three, including
`(*SagaCoordinator).applyResponseToState`, 4,746 chars. ⚠ Note it is stored **receiver-qualified**;
a `code_request` naming `applyResponseToState` bare returns nothing. Now a LANDMINE.)

### But the verdict is UNVERIFIABLE again — for a NEW and simpler reason

`stopped: iteration-cap`, **not** `scope-not-narrowing`. It ran five full iterations and searched
properly. It found nothing because **the rows this bug rests on no longer exist:**

```sql
SELECT count(*) FILTER (WHERE collected_data ? 'hero_deployed') AS hero,
       count(*) FILTER (WHERE collected_data ? 'logo_deployed') AS logo,
       count(*) AS retained
FROM orchestration_states;
--  hero | logo | retained
--     0 |    0 |     2110        (2026-08-11 10:15Z)

SELECT count(*) FROM orchestration_states
WHERE orchestration_id = '3e46be5b-8788-447b-9643-e32ae33f601b';
--  0     -- §1's DECISIVE ROW IS GONE
```

**This is the third state this bug has been observed in**, and the pattern is now the finding:
0 of 1,667 on 08-06 → 2 each on 08-09 (§1) → **0 again on 08-11**. The keys appear only while a
site build that deploys a hero/logo is inside its retention window. `2,110` rows are retained and
the oldest is 2026-07-13, so this is **not** a flat 24h TTL on the table — these particular
orchestrations are gone, which is a different question from how long the table keeps rows.
`[UNVERIFIED]` which reaper or prune path removed them.

### What this bug needs now — it is no longer a search problem

**Re-running `090` again will fail the same way and cost credits for nothing.** The loop is no
longer blind; there is simply nothing left to look at. Two ways forward, and they are not equal:

1. **Capture at the moment of failure (recommended).** Commission **item 2** — making the three
   silent readers (`v3_site_actions.go:1020`, `:1031`, `assemble_from_library.go:452`) log the key
   they wanted **and the keys the map actually held** — writes exactly this evidence into
   `agent_error_log` when it next happens, where nothing prunes it on an orchestration's schedule.
   **Item 2 is therefore the unblocker for this bug, which the commission's ordering did not
   anticipate.**
2. **Force a fresh occurrence** by running a site build that deploys a hero/logo, then reading
   `collected_data` while the row is still live. Faster if something is already queued, but it is
   a race against the same prune that just erased the last one.

**Do not treat the `(0 rows)` above as evidence about the mechanism.** It is evidence about
retention. The §5 refutation and the candidate directions there are untouched by this run.

---

## CONTRIBUTION 2026-08-11 13:35Z — why the evidence keeps evaporating: the window is **4 hours**, not 24, and not 7 days

Left by the `bugfix_236_site_availability` lane (the *other* 236 — the 522 one).
**Not taking this bug.** `who-owns` shows this file committed against as recently
as 11:21Z today (`e41342c89`) and four live sessions reference `hero_deployed`, so
this is contribution, not competition. It answers exactly one open question — the
one `e41342c89` explicitly left: *"pruned by something I did not chase"*.

**The pruner is `database-cleanup`** (`scheduled_tasks`, enabled, hourly, last ran
2026-08-11 12:04:11Z). Read from the **live row**, not the repo seed, and that
distinction is load-bearing here — they disagree:

| status | live retention | what `020_scheduled_tasks.sql:739` says |
|---|---|---|
| `COMPLETED`, `FAILED` | **24 hours** | 7 days |
| `EXECUTING_STEP`, **`AWAITING_RESPONSES`** | **4 hours** | 24 hours (and names `WAITING_FOR_RESPONSE`, a status that does not occur) |

```sql
SELECT pre_query FROM scheduled_tasks WHERE name='database-cleanup';
--   DELETE FROM orchestration_states
--    WHERE status IN ('COMPLETED','FAILED')  AND updated_at < NOW() - INTERVAL '24 hours'
--   DELETE FROM orchestration_states
--    WHERE status IN ('EXECUTING_STEP','AWAITING_RESPONSES') AND updated_at < NOW() - INTERVAL '4 hours'
```

**Why this is the answer and not a coincidence:** a `site-work-orchestrator`
waiting on a spawned child sits in **`AWAITING_RESPONSES`** — which is precisely
the state in which `hero_deployed` / `logo_deployed` exist, because they ARE the
awaited responses. So the rows this bug needs are in the **shortest-lived
category in the table**. The three states you have observed —
`0 of 1,667` → `2 each` → `0` — are not the bug appearing and disappearing. They
are one 4-hour window opening and closing.

> **CORRECTION to `e41342c89`, offered with the evidence.** That commit reasons:
> *"NOT a 24h TTL claim: 2,110 rows are retained and the oldest is 2026-07-13."*
> The premise is true and the conclusion does not follow, because the population
> is mixed. Measured 2026-08-11 12:30Z:
>
> | status | rows | oldest |
> |---|---|---|
> | COMPLETED | 2,329 | **2026-08-10** |
> | FAILED | 23 | 2026-08-10 |
> | EXECUTING_STEP | 6 | 2026-08-11 |
> | CANCELLED | 24 | 2026-07-19 |
> | RUNNING | 20 | 2026-07-29 |
> | INITIALIZED | 1 | **2026-07-13** |
>
> The 2026-07-13 row is a single `INITIALIZED` orphan. `CANCELLED`, `RUNNING` and
> `INITIALIZED` are named by **no arm of the pruner**, so they never expire — they
> are a small permanent leak, and they are the entire reason the table *looks* like
> it has a month of history. The populations that matter here are one day
> (`COMPLETED`) and four hours (`AWAITING_RESPONSES`). **A `min()` over a mixed
> population measured the leak, not the retention.**

### What this changes for whoever fixes it

1. **Stop trying to catch the row by looking.** With a 4-hour window and an hourly
   pruner, a query run the next morning is guaranteed to find 0. That is not
   evidence of absence and never was.
2. **Capture, don't query.** Snapshot the shape into a durable table the moment a
   build runs, e.g. `CREATE TABLE bug236_capture AS SELECT * FROM
   orchestration_states WHERE collected_data ? 'hero_deployed'` inside a poll that
   runs every few minutes during a build, or add a `log_action_findings` call on
   the reader step. The decisive row `3e46be5b` is gone and will not come back.
3. **A 090 run against this table inherits the same window.** `e41342c89` records
   the verdict as UNVERIFIABLE with the stop reason changed to iteration-cap, and
   reads that as the loop "searching and finding nothing". That is right, and now
   it has a mechanism: the loop could not have found the rows *whenever it ran more
   than four hours after the build*, regardless of how well it searched.
4. **`WAITING_FOR_RESPONSE` (singular) appears in the repo seed and occurs in no
   live row.** If anything else greps the seed for orchestration statuses, it is
   matching a status this system does not emit.

**Not verified by me:** anything about the `image_url` loss itself. I did not read
the readers, and this contribution makes no claim about §5's open root cause.

---

## CONTRIBUTION 2026-08-11 (later) — item 2 is BUILT, and a §5 candidate that explains BOTH missing-key observations

Left by the `silent_hero_logo_readers` lane (commission item 2). Two parts: what shipped, and a
lead that is **NOT** a root-cause claim.

### Part 1 — the readers are no longer silent (commission item 2, owner "2. yes.")

All three sites now route through `deployedImageURL`
(`platform/orchestration/actions/deployed_image_read_audit.go`). On a container that is **present
but carries no usable `image_url`** they emit a `Warn` **and** record one `agent_error_log` row:

| | |
|---|---|
| `error_code` | `DEPLOYED_IMAGE_RESULT_MISSING_URL` |
| `severity` | `warning` |
| context | `container_key`, `wanted_key`, **`keys_present`** (sorted; keys only, never values), `container_type`, **`fallback_sibling_present`**, `remedy` |

**Why a durable row and not only the commissioned `Warn`:** the CONTRIBUTION above measured the
window at 4 hours, and `agent-chassis` does not retain a log line long enough either (its own
startup line was measured absent from `--tail=3000` hours after a roll). `agent_error_log` is
documented as the only sink that outlives an awaited step. So this stops *querying* for the shape
and starts *capturing* it — which is what item 2 of the CONTRIBUTION above asked for.

**Demand-gated:** an ABSENT container records nothing, because most pages deploy no hero or logo.
Only present-but-unusable records. Proven by mutation.

**`fallback_sibling_present` is aimed squarely at §5** — see Part 2 for why it is the
discriminator, and note the standing query in the lane RUNBOOK needs a **demand control** beside
it, per §3's warning against reading a bare count as an incidence rate.

⚠ Live on the next fleet roll; the code is at HEAD (swept into `038211dd8`, verified restored and
green from a clean `git archive HEAD`). Lane: `docs024_key_docs_latest/silent_hero_logo_readers/`.
Council `Council-Submitted: c80ea1d7-ce1e-493f-8175-877501d895e6`.

### Part 2 — a candidate for §5, offered as a CANDIDATE and marked as one

**This is not a root-cause claim and must not be quoted as one.** It belongs on §5's
`[UNVERIFIED]` candidate list, alongside the four already there. What I did first-hand is read
three functions; what I did **not** do is observe it happen.

**The mechanism: the park path writes a state it re-loaded from the DB, and never copies
`CollectedData` onto it.** `persistAwaitingStateWithRetry` (`coordinator.go:2067-2132`):

```go
freshState, err := repo.GetState(ctx, state.OrchestrationID)   // :2073 — from the DB
...
for k, v := range state.AwaitedRequests {                      // :2093 — ONLY awaited requests
    freshState.AwaitedRequests[k] = v
}
freshState.Status = StatusAwaitingResponses                    // :2098
freshState.LastActivity = time.Now()                           // :2099
err = repo.UpdateState(ctx, freshState)                        // :2102 — freshState is what lands
```

`state.CollectedData` is **never copied onto `freshState`**, so the row that gets written carries
whatever `CollectedData` was in the DB *before* the action ran.

**And nothing persists in between.** `storeActionResult` (`:1863`, called at `:1795`) mutates
`state.CollectedData` **in memory only**; the next `repo.UpdateState` in that path is the one
inside the park at `:2102`, reached via `processAwaitResponse` at `:1839` → `:2000`. Checked by
enumerating every `UpdateState` call site in the file: between `:1795` and `:2000` there are none.

**This is not a new theory — it is the one this codebase already states, applied here.**
`agenterrors.go:20-24`, verbatim: *"the sibling collected_data key was REFUTED live
(persistAwaitingStateWithRetry loads fresh state at park time and copies only awaited-request
entries across; RFC_012 addendum 2). THIS TABLE IS THE ONLY SINK THAT SURVIVES AN AWAITED STEP."*

**Why it is worth adding to §5: it explains BOTH observations, which no candidate there does.**

1. `hero_deployed`/`logo_deployed` lack `image_url`/`output_path`/`size_bytes` — the keys
   `deploy_image_asset_action.go:369-371` assigns to its own result, in memory, before the park.
2. `collected_data` lacks `hero_url`/`logo_url` (§1's second query) — **despite
   `deploy_image_asset_action.go:404-415` having written exactly those sibling keys since
   `d45c86b1e`, 2026-02-23.** That workaround's own comment says it exists *"so it survives the git
   adapter response overwriting this step's output_field"*. It is the same in-memory
   `CollectedData` mutation, so on this reading it is discarded by the same park — which is why a
   five-month-old fix has never worked, and why nobody noticed.

**It also leaves §5's refutation intact.** The merge at `:2719-2748` does preserve-then-add, exactly
as §5 says. On this candidate the merge is innocent: it preserves a map that never carried the
action's keys in the first place, because the loss happened one step earlier, at park.

**What would confirm or kill it** (none of this is done):

- The **cheap kill**: if `collected_data` on a parked row ever carries `image_url` under
  `hero_deployed`, the candidate is wrong. `[UNVERIFIED]`
- The **direct test**: does any live row show an action's own pre-await result keys surviving the
  park at all? That is a general question about every action that mutates `CollectedData` before
  awaiting — the blast radius is the reason this deserves the loop rather than my confidence.
- **Item 2's new rows will answer it forward.** `fallback_sibling_present` is precisely this
  discriminator: **false** on a real occurrence means the URL is gone from the container *and* from
  the sibling key, which is what this candidate predicts and what candidate 2 in §4 would not.

> **Filed to the diagnosis loop rather than asserted**, per CLAUDE.md's rule for a cause that lives
> outside the symptom and a claim whose blast radius is fleet-wide. **If that verdict contradicts
> this, the verdict wins** — a refuted candidate costs one run, and this one is written to be
> disconfirmable.
>
> **`090` FILED 2026-08-11.** `RUN_CORRELATION_ID=dbcc4259-ab84-494b-a48b-1df647209a40`
> (intake `7521cfee-73ed-4fe4-a653-91fee2ded5f4`; the RUN id is the one the artifacts are written
> under). The symptom names the mechanism and points at the four functions and at
> `orchestration_states.collected_data`; it asserts **no rows and no counts**, so the loop fetches
> and cites its own. **VERDICT NOT YET READ — nobody should treat this candidate as confirmed
> until it is.**
>
> ⚠ Unlike the two earlier `090` runs on this bug, this one does **not** depend on the 4-hour
> evidence window: it asks a question about a **code path**, which is why it is worth its credits
> where a third re-run of §5b's question would not be.

### `090` VERDICT READ 2026-08-12 — UNVERIFIABLE. The candidate is NEITHER confirmed NOR refuted

Run `dbcc4259-ab84-494b-a48b-1df647209a40`, COMPLETED 2026-08-11 18:42Z, 4 bundle iterations.
**Do not quote this as a refutation** — and do not quote it as support either. Its own
`needed_evidence`, verbatim:

> *"The bundle never renders the bodies of `persistAwaitingStateWithRetry`, `processAwaitResponse`,
> or `applyResponseToState` — only `storeActionResult`'s body and a bare signature line for
> `applyResponseToState` are present. Without the actual copy logic inside
> `persistAwaitingStateWithRetry` … there is no static evidence either confirming or refuting the
> claimed field-list ('only AwaitedRequests, Status and LastActivity')."*

Plus a state-tier miss: *"the diagnosis target correlation_id has no orchestration_states row at
all"*, and its own `data_request` for a parked row carrying `image_url` returned unrelated rows —
because **nothing has deployed a hero or logo since** (see the demand control below).

It reached the same `next_scope` I did (`persistAwaitingStateWithRetry`, `processAwaitResponse`,
`applyResponseToState`, `DeployImageAssetAction`) and cited `storeActionResult`'s
`state.CollectedData[state.CurrentStep] = result`. So it agrees on where to look and could not look.

### ⚠ THE BLOCKER IS THE BUNDLE, NOT THE INDEX — and §5b (mine) checked the wrong thing

§5a item 1 recorded this same "could not read the function bodies" blocker. **§5b declared it
clear**, on this evidence: *"the index is fresh and carries all three, including
`(*SagaCoordinator).applyResponseToState`, 4,746 chars."* That was true, and it was **not an
answer to the question that mattered.**

Measured 2026-08-12 against `code_symbols` — every body the loop said it lacked **is present**:

| symbol | kind | `length(body)` | lines |
|---|---|---|---|
| `persistAwaitingStateWithRetry` | func | **2,058** | 2067–2132 |
| `processAwaitResponse` | func | **5,619** | 1914–2063 |
| `(*SagaCoordinator).applyResponseToState` | method | **4,746** | 2650–2779 |
| `storeActionResult` | func | **970** | 1863–1892 |

**So the index held all four and the bundle rendered one.** The defect is in the code tier's
selection/rendering, not in indexing — which means §5b's check *could not have come out false* for
the loop's actual failure mode, the shape `WRONG_CALLS.md` exists to record. Logged there.

**This closes an open question commission item 5 explicitly left.** That lane's PLAN §3 said:
*"`code_symbols` is now described in the schema section, but whether the code tier has an analogous
blind spot is unexamined. `[UNMEASURED]`"* — it is measured now, and the answer is **yes, the same
shape**: item 5 fixed the SCHEMA tier's silent filtering; the CODE tier drops bodies it holds, and
the verdict's cite-or-abstain rule then acts on that absence exactly as it did for the schema.

**Consequence for this bug: a third `090` on the code-path question will fail the same way until the
code tier is fixed.** Route it as a diagnosis-harness defect first, not as another run here.

### What the roll changed, and what it did not

| check | result | date |
|---|---|---|
| item 2 live on the chassis | **YES** — `v1.0.1290`, `DEPLOYED_IMAGE_RESULT_MISSING_URL` present in `/proc/1/exe` on **both** replicas, negative control absent | 2026-08-12 |
| `agent_error_log` rows with that code | **0** | 2026-08-12 |
| **DEMAND CONTROL** — anything deploying a hero/logo at all | **`hero_deployed` 0, `logo_deployed` 0** of 6,364 retained | 2026-08-12 |

**The zero is therefore unfalsifiable, not reassuring.** There has been no demand on the path since
the roll, so item 2 has not yet had an opportunity to fire. Per §3's own warning, do not read this
count as an incidence rate in either direction. The behavioural proof still needs a site build that
deploys a hero or logo.

---

## CONTRIBUTION 2026-08-12 (evening) — §5's harness blocker is GONE; the loop has read the code, and its first partial finding CONTRADICTS my §5b hypothesis

**Three `090` runs on §5's question, and the third is the first to read any of the code.**

`bugs_closed/261` (the code tier could not resolve receiver-qualified method names) is fixed and
live on `v1.0.1293`. That was the thing making runs 1 and 2 useless. It is no longer the blocker.

Run `36bd1b42-29b5-4094-9264-94ea80c6194a`, seeded with the four functions via `SEED_SCOPE`
(the seed **arrived** — no `bugs_open/174` confiscation; checked, not assumed).
**Iteration 1 rendered all four bodies.** Verdict still `UNVERIFIABLE`, but for the first time it
carries **citations**:

| tier | where | quote |
|---|---|---|
| static | `persistAwaitingStateWithRetry` | `if existingData, exists := freshState.CollectedData[state.AwaitedRequests[reqID].StepName].(map[string]interface{}); exists {` |
| static | `persistAwaitingStateWithRetry` | `if freshState.AwaitedRequests == nil {` |
| static | `(*SagaCoordinator).applyResponseToState` | `state.CollectedData[stepName] = normalisedData` |

### ⚠ THE HYPOTHESIS IN §5b MAY BE WRONG, AND IT IS MINE

The loop's own `needed_evidence`, verbatim:

> *"one of them ('if existingData, exists := freshState.CollectedData[...]') already shows the
> function DOES reference freshState.CollectedData, **which is inconsistent with the hypothesis's
> literal claim that it 'copies only AwaitedRequests, Status and LastActivity onto it'** — but the
> fragment is an existence-check inside a loop, not the assignment logic, so it's unclear whether
> the in-memory state.CollectedData … is actually merged onto freshState before repo.UpdateState."*

**This is not a refutation and must not be recorded as one.** It is one fragment, and the loop says
so itself. But §5b's claim was inherited from `agenterrors.go`'s comment (*"copies only
awaited-request entries across"*) and repeated by me without reading the function. The one line we
now have suggests the real behaviour is a **merge**, not a wholesale replace — which would mean the
mechanism §5 proposes for the lost keys is wrong, and the true cause is elsewhere.

**Do not fix anything on the strength of §5b until this is settled.** Marked `[CONTESTED]`.

### What still blocks it — and it is no longer the code tier

1. **The loop RE-SCOPED AWAY from its own key function.** Iteration 1 held all four bodies; by
   iteration 5 `persistAwaitingStateWithRetry` and `processAwaitResponse` were **not in scope at
   all** (18 symbols across 8 files, neither of them present). **The bundle does not accumulate and
   the verdict reads only the LAST iteration**, so the seeded evidence was gone by the time the
   verdict was written. Seeding fixes iteration 1 and nothing after it. *This is now the biggest
   single obstacle to answering §5.*
2. **No state-tier evidence exists to be found.** The verdict needs a row parked mid
   `deploy_image_asset` (`status='AWAITING_RESPONSES'`) and there is none — same demand drought that
   makes item 2's zero unfalsifiable (§3). The three parked rows it did find were `run_triage`,
   `process_sites_iter_3_call_orchestrator`, `call_dispatch`.
3. `bugs_open/267` — an iteration can be spent on a whole-file re-read the bundle advises and its own
   arithmetic refutes. Cost this run one of five iterations.
4. One orchestration row **FAILED** at `call_diagnoser`: *"Request ee294db7-9c25-43ed-b45c-5c3e54cee8be
   timed out after 3 retries"*, `__step_error` empty. Not investigated.

### The cheapest next move

**Re-run seeded with ONLY `persistAwaitingStateWithRetry`** (2,058 chars — no cap risk, no room to
re-scope away from it) and a symptom that asks the narrow question: *what does it copy onto
freshState before UpdateState?* One function, one question. The broad symptom is what let the loop
wander across eight files.

The state-tier half cannot be forced and should not be waited for: catch it opportunistically the
next time anything deploys an image.

---

## CONTRIBUTION 2026-08-14 — the "cheapest next move" was never actually run; its 08-12 re-dispatch FAILED with no verdict; the narrow run is now dispatched

Left by the `silent_hero_logo_readers` lane, picking this up per its handoff.

**All three tooling blockers this file kept hitting are now retired.** `bugs_closed/261` (reader),
`bugs_closed/267` (iteration waste) and `bugs_closed/269` (sibling handles ambiguous — closed today,
live on chassis `v1.0.1297`, 12/12 sibling method handles canonical in the first post-roll bundle).

**What the section above prescribed — seed ONLY `persistAwaitingStateWithRetry` — never ran.** The
coverage check on today's dispatch surfaced it: a third `needs_diagnosis` item
(`686f58a1-2431-42f0-98bc-2e0537069c2c`, created 2026-08-12 19:49:59Z, ~2 minutes after run
`36bd1b42`'s item completed) re-used the **four-function seed and the broad symptom verbatim**, and
**failed**:

- item status `failed` at 20:30:50Z;
- five bundles written under its `dispatch_correlation_id` (`36bd1b42…` — it reuses the earlier
  run's key, so those bundles interleave with that run's history) from 19:56:50Z to **20:33:21Z —
  the last one AFTER the item was already marked failed**;
- **no verdict anywhere**, and the orchestration row is pruned, so the cause is `[UNVERIFIED]` and
  now unrecoverable. Filed here as harness evidence, not diagnosed.

**A retrieval fact worth keeping: verdicts are NOT written to `diagnosis_artifacts`.** For all three
runs on this question (`90f6f55f`, `dbcc4259`, `36bd1b42`) that table holds **bundles only** (the
kinds present at all: `bundle`, `iteration_note`, `fix_plan`, `council_report`, `escalation`). A
verdict must be read from the run's orchestration row **within its retention window** — which is
also why the failed run's cause is gone.

**Dispatched today, exactly per the prescription above** (FORCE=1 after reading the coverage hit —
the blocking item was this lane's own failed 08-12 run, no live session on it; transcripts checked):

| | |
|---|---|
| intake correlation | `7daa0c43-4b1c-40f4-8f7a-7aa7817b3251` |
| **RUN correlation (artifact key)** | **`23f1cf9a-2e33-43a3-9b33-d18adbbe5c55`** |
| seed | `platform/orchestration/coordinator.go:persistAwaitingStateWithRetry` — one function, no cap risk, no room to re-scope away |
| symptom | the one narrow question: which fields does it copy onto `freshState` before `repo.UpdateState`; is `state.CollectedData` merged or left behind |

**VERDICT NOT YET READ.** Whoever reads it: do so within the orchestration row's retention window,
and record it here either way — §5b stays `[CONTESTED]` until then.

### VERDICT READ 2026-08-14 ~08:10Z — the narrow question is ANSWERED, the `[CONTESTED]` marker is lifted, and the mechanism is now CONFIRMED with a declared first-hand supplement

Run `23f1cf9a…`, three bundle iterations, verdict `UNVERIFIABLE` — but read what it says rather than
the label, because this one is different in kind from the four before it:

> *"The full body of persistAwaitingStateWithRetry (given in full in this bundle) copies only
> AwaitedRequests, Status, and LastActivity from the in-memory `state` onto the freshly-loaded
> `freshState` before UpdateState — the cited existence-check reads freshState's OWN CollectedData
> (to detect an already-arrived response), it does not merge state.CollectedData onto freshState,
> so it does not contest the plain-discard reading; if anything it reinforces it."*

**So the fragment that made §5b `[CONTESTED]` is resolved: it was an existence check, not a merge.**
The loop finally read the whole function (the one-function seed left it no room to re-scope away)
and its five citations include both halves of the copy logic. The `UNVERIFIABLE` label is only the
confirm-needs-occurrence rule: static mechanism established, live occurrence not yet witnessed by
the loop — its own two `data_requests` came back truncated before the deciding key was visible.

**The two residual checks it enumerated were run first-hand the same hour, and this is the declared
substitution per the 2026-07-31 owner ruling** (the loop named exactly what would settle it; these
are those checks, not a fresh theory):

1. **Occurrence — its own two data_requests, untruncated, while the rows were still parked.** Both
   were this very run's OWN children, sitting in `AWAITING_RESPONSES` inside the 4-hour window:

   | orchestration | parked on step | step's key in `collected_data` | earlier steps' keys |
   |---|---|---|---|
   | `74f8683d…` | `call_diagnoser` | **ABSENT** | `spawn_diagnoser` etc. present |
   | `5622ffcc…` | `call_handler` | **ABSENT** | `spawn_handler`, `handler_spawned` etc. present |

   Every completed step's key survives; the currently-awaited step's key — the one
   `storeActionResult` wrote in memory moments before the park — is gone from the persisted row.
   Both rows, same shape.

2. **Ordering — `processActionResult`'s body, read in the tree** (the loop had located only the
   `processAwaitResponse` call site): `coordinator.go:1795` calls `storeActionResult`
   unconditionally FIRST — which writes `state.CollectedData[state.CurrentStep] = result` (`:1873`)
   and `state.CollectedData[step.OutputField] = result` (`:1877`), in memory — and `:1839` calls
   `processAwaitResponse` (→ the park) AFTER, with the same `state`. So for every awaited step the
   in-memory write demonstrably precedes the discard; "the key is only written at response time"
   is eliminated as an innocent explanation for the absences above.

**What is CONFIRMED:** the park path (`persistAwaitingStateWithRetry`) discards every
`CollectedData` mutation made in memory during the dispatching step — mechanism loop-cited from the
full body, occurrence witnessed live on two independent parked rows, ordering verified at
`:1795`/`:1839`. This is fleet-wide: every action that enriches its own result (or writes sibling
keys) before an await loses that work at park; the response-time merge then "preserves" a map that
never held it. It explains both §1 observations, including why the five-month-old sibling-key
workaround (`deploy_image_asset_action.go:404-415`, `d45c86b1e`) has never worked.

**What remains `[INFERRED]`:** that this mechanism is what erased THESE hero/logo keys on 08-09 —
those rows are pruned and no direct witness exists. The mechanism is confirmed in general;
item 2's `agent_error_log` capture (`fallback_sibling_present=false`) will witness the hero/logo
instance the next time anything deploys an image.

**What this bug now needs is a DECISION, not another run:** the fix lands in the awaited-park seam
(merge `state.CollectedData` onto `freshState` at park, or persist before parking) — which is
RFC_012 `(a)`/`(a′)` territory, an open OWNER call, with every await-using pipeline in the blast
radius. Do not patch it inside this bug; route it to that decision with this evidence attached.

---

## CONTRIBUTION 2026-08-15 — the park fix is RULED and BUILT; and two findings that change what this bug IS

By the `silent_hero_logo_readers` lane. **The owner ruled RFC_012 (a) for the park path,
additively** (RFC_012, third sitting). The fix is in the tree, tested and mutation-proven; it is
Go, so it is inert until the next chassis roll. Register entry **WFA-014**.

**What the fix does for this bug:** the park no longer discards the dispatching step's in-memory
`collected_data`, so `deploy_image_asset`'s own `image_url` / `output_path` / `size_bytes` now
reach the persisted row instead of dying at `persistAwaitingStateWithRetry`. That is the
mechanism §5/§5b spent five contributions narrowing and the 08-14 run confirmed.

### Finding 1 — hero and logo are NOT the same case, and this file's title implies they are

Verified 2026-08-15 against live `agent_definitions`:

| step | `output_mapping`? | reply-time branch | where `image_url` comes from |
|---|---|---|---|
| `deploy_hero_image` (both `site-work-orchestrator` and `pageflow-builder`) | **yes** | `applyOutputMapping` — **replaces the record wholesale** | the REPLY, via `"image_url": "response.data.file_path"` |
| `deploy_logo_image` (both) | **no** | `.response` branch — **additive, preserves the map** | the ACTION's own key, which the park was discarding |

So the park fix addresses the **logo**. For the hero the mapped result replaces everything at
reply time, carried keys included, and its `image_url` is supposed to arrive from the reply
independently.

**This is consistent with §1's own evidence, which is worth re-reading in that light:** the
decisive non-timeout row (`3e46be5b…`) is a **logo** row, and its persisted shape —
`{response:{…}, response_status, response_received_at}` and nothing else — is exactly what the
additive branch produces over a map the park had emptied. The two hero rows in §1 were timeouts
and §1 already excludes them as evidence. **[UNMEASURED]** whether the hero's mapping path
resolves in practice; nothing here shows it failing, and nothing here shows it working.

### Finding 2 — the readers cannot see these keys AT ALL, and this is a separate, larger defect

The three readers §1 cites have since been collapsed onto one helper (`deployedImageURL`,
`deployed_image_read_audit.go:114`) — so §1's `v3_site_actions.go:1020`/`:1031` line numbers are
stale; the live call sites are `v3_site_actions.go:1187`/`:1196` and
`assemble_from_library.go:456`.

Those readers run in **`page-content-writer`** and the **architect** agents. The producing steps
run in **`site-work-orchestrator`** and **`pageflow-builder`**. Different agents means different
orchestrations and different `collected_data`. Verified 2026-08-15: **the only references to
`hero_deployed` or `logo_deployed` anywhere in live config are the four producing steps** — no
`input_mapping`, no template reference, nothing forwards them. A spawned child receives
`input_data`, not the parent's map (`spawn_actions.go`).

**So the readers hit their own "NO DEMAND" early return every time, and file nothing.** That
predicts a zero from the instrument built on 08-11 to size this bug — and the zero is there:

```sql
SELECT count(*) FROM agent_error_log WHERE error_code='DEPLOYED_IMAGE_RESULT_MISSING_URL';
-- 0, against a live table (9,288 rows / 26 distinct codes since 2026-08-11)
-- demand control: 7 needs_hero_image + 187 undeployed_asset completions in the same window
```

**A zero with real demand behind it is not "the bug stopped happening" — here it is the detector
being blind in the exact shape the confirmed mechanism produces.** The instrument's own
early-return comment calls the absent-container case "NO DEMAND"; for these keys the container is
absent *structurally*, not because nothing deployed an image.

**What this means for the fix:** the park fix is correct and lands a real platform defect, but
**it will not by itself put a hero or logo URL on a page**, because the code that would consume
the restored key never receives the key. Do not close this bug on the roll. The remaining
question is a routing one — whether these keys should cross to the reading agents, or whether the
readers are vestigial and the live path is the `site_assets` resolver — and it wants an owner
decision or its own bug, not a patch. **[INFERRED]** that the readers are the intended consumer
at all; that assumption has been carried since 08-09 and has never been tested.
