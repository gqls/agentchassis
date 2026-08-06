# 183 — `domain-research-classifier` truncates on a 6000-token cap, the fleet's only one, and burns all 3 attempts

**Filed 2026-08-03** by the `mortgagecalculator_couk_adoption` lane, which it blocked.
**Status: cap raised 6000 → 16000, then → 32000 on the owner's instruction
(live config; 16000 proven in production first — one run at 6590 output tokens).
OPEN** because the exposure that produced it is untouched — see fix candidates 2 and 3.

> **UPDATE 2026-08-06 — candidate 4 is DONE and live; this bug is now one decision
> from closing.** The standing check shipped (see candidate 4 below) and found a new
> live bug on its first run (`bugs_open/205`). What remains OPEN here is a single
> question, and it is a judgement, not an investigation: **candidate 3** — split
> `classify_and_extract`'s four documents into four bounded generations. It is
> architecture-scope and deliberately not attempted while adoption lanes are using
> this agent. With the cap at 32000 the observed max (6590) has ~5× headroom, and
> regrowth toward that ceiling is now ANNOUNCED rather than discovered by a burned
> site. **The honest closing condition:** close this when the owner rules that
> headroom-plus-monitoring is the accepted answer for this step; keep it open if
> candidate 3 is still wanted. Do not close it on the cap raise alone — that was
> true on 08-03 and the class exposure was still there.

> **UPDATE 2026-08-03 — this is NOT a one-step problem, and the fleet number is
> much worse than this bug's.** A 14-day audit keyed on `error_message` (the only
> reliable truncation signal) shows **25 distinct steps truncating**, and ours is
> mid-table. The council review seats are the real concentration, at cap 8000:
> `review_editquality` **21/105** and **19/136** on two agent_types, with a maximum
> successful call of **7996 of 8000 — 99.95% of cap**; `review_contracts` **6 of
> 10**; `review_feasibility` **5 of 10**. Two `experience-approval-council` seats
> sit on the same 6000 this bug is about.
> **Left alone deliberately.** Those seats are `bugs_open/138`'s active lane, which
> has already raised `editquality` once and recorded that it *grew back into the
> doubled cap in three days* — so a second unilateral raise from a passing session
> is both territorial trespass and the move that lane has already shown does not
> hold. Raising twenty caps mid-flight would also disrupt councils other sessions
> are running right now. Reported to the owner with the numbers instead.

## Symptom

A newly adopted site gets **no strategic specs at all**. The work item
`needs_domain_research` (handler `domain-research-classifier`) goes `failed` with:

```
step classify_and_extract failed: failed to execute action execute_llm_prompt:
AI call failed with unhandled error: response truncated: stop_reason=max_tokens
(output_tokens=6000 reached the configured cap, 26179 chars recovered);
raise max_tokens or shorten the prompt (code: CHILD_ORCHESTRATION_FAILED)
```

`classify_and_extract` is step 6 of 15 and precedes **all four** `write_*_spec`
steps (`write_identity_spec`, `write_classification_spec`,
`write_content_direction_spec`, `write_design_intent_spec`). So the failure is
clean — it writes nothing, corrupts nothing — but the site is left with only its
adoption-seeded specs and never gets classified.

`max_attempts` is 3, and the failure is near-deterministic for a given site, so a
hit **burns all three attempts** and the item is dead without operator action.

## Evidence

**The step has run 54 times since 2026-04-02. Zero truncations until 2026-08-02;
5 of 6 that day.** Cap and model are constant across the whole window
(`model_resolved` identical on every row — not an alias drift).

```sql
SELECT created_at::date AS day, count(*) AS calls,
       count(*) FILTER (WHERE error_message ILIKE '%stop_reason=max_tokens%') AS truncated
  FROM llm_call_log WHERE step_name='classify_and_extract'
 GROUP BY 1 ORDER BY 1 DESC;
--  2026-08-02 | 6 | 5      <-- all of them
--  every prior day        | 1-7 | 0
```

**The cap has always had ~6% headroom above the observed maximum**, measured over
the 49 successful calls:

| stat | tokens | % of the 6000 cap |
|---|---|---|
| mean | 4592 | 76.5% |
| p95 | 5551 | **92.5%** |
| max (2026-07-17) | 5642 | **94.0%** |

The one call that survived 2026-08-02 landed at 5602 (93.4%). This was never a
regression — it was a latent failure with a two-token-wide margin, and the
distribution's tail finally crossed it.

**6000 is the ONLY step in the fleet at that cap.** Across every active,
non-snapshot agent definition:

```sql
WITH steps AS (SELECT ad.type, s.key AS step_name,
       (s.value->'config'->'ai_service'->>'max_tokens')::int AS cap
  FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
 WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
SELECT cap, count(*) FROM steps WHERE cap IS NOT NULL GROUP BY 1 ORDER BY 1;
-- 4000:10  6000:1 (this one)  8000:47  16000:20  32000:12  64000:1
```

The two modes are **8000 (47 steps)** and **16000 (20 steps)**. This step emits one
of the largest documents in the estate — a four-section JSON whose own prompt sets
floors of ≥8 `writing_rules`, ≥6 `things_to_avoid`, ≥6 `things_to_emulate`, 4–10
`industry_tags`, 5–10 `key_terms`, and 8 mandatory palette hex slots — while
carrying the lowest cap above 4000.

## What is NOT the cause — measured, not assumed

- **NOT the adoption branch.** The prompt has a large `{{if .site_specs.specs.site_archetype}}`
  "Adoption Reference" block, and the obvious theory was that adopted sites echo
  back more and so overrun. **Refuted.** Splitting all 54 calls on
  `prompt_rendered LIKE '%Adoption Reference%'`: non-adoption **10.0%** truncated
  (2/20), adoption **8.8%** (3/34). Both classes truncate and both classes only
  truncate on 2026-08-02.
- **NOT a model or alias change.** `model` and `model_resolved` are
  `claude-sonnet-4-6` on all 54 rows.
- **NOT the definition edit of 2026-08-02 22:08.** That timestamp bumped **184
  agent_definitions rows** — a bulk sweep — and it lands ~8 hours *after* the
  failures (13:31–14:04 UTC). It did not change this step's cap, model or prompt.
- **NOT invisible thinking eating the budget.** 26179 chars recovered at 6000
  output tokens is 4.36 chars/token, normal for JSON; the visible document really
  is that long. (Contrast the trap in `MEMORY` / bugs_open/138, where a cut call
  showed 42 visible chars at cap 120.)

> **[UNEXPLAINED] Why the tail crossed on 2026-08-02 specifically.** Cap, model and
> prompt were all unchanged, and the one structural hypothesis was refuted above.
> The defensible claim is the margin, not a trigger: a step whose p95 sits at 92.5%
> of its cap will truncate, and asking *which* input tipped it is the wrong
> question. Do **not** repeat "a change on 2026-08-02 caused this" — nothing has
> been found that did, and the low-headroom finding does not need one.

## Root cause

The cap was sized for an earlier, smaller version of this prompt and never
revisited as the prompt grew (it now renders to ~17,000 chars of template alone).
The step has no truncation tolerance — `execute_llm_prompt` surfaces
`*aiservice.TruncatedError` as an unhandled error — so a cut is fatal to the step,
the child orchestration, and (after 3 attempts) the work item.

## Fix candidates, ordered by what closes the door

1. **[APPLIED] Normalise the outlier cap: 6000 → 16000.** Live DB config, no image
   rebuild. Verified live and verified *not shadowed*:

   ```sql
   -- bugs_open/009 interaction: a ROOT ai_service block makes step-level max_tokens
   -- dead config. This agent has NO root block, so the step value is the live one.
   SELECT (default_config #> '{ai_service}') AS root_block,
          default_config #>> '{workflow,steps,classify_and_extract,config,ai_service,max_tokens}'
     FROM agent_definitions WHERE type='domain-research-classifier' AND is_active;
   -- root_block = NULL, step = 16000
   ```

   Independently corroborated: every pre-change `llm_call_log` row recorded
   `max_tokens=6000`, exactly the step value, so the step block was always the live one.
   **This buys headroom; it does not close the door** — see `platform/aiservice/truncation.go:26-29`,
   which says so in the platform's own words ("experience-planner/compose truncated at a
   32,000-token cap… whatever the number, the step that writes most approaches it").

2. **Do NOT add a `repairTruncatedJSON` salvage path to this step.** It exists
   (`apply_adoption_plan_action.go:1011`) and is the right answer for the councils,
   but it is the **wrong** answer here and the reason is worth writing down: the
   repair keeps a prefix ending at a complete value, so **trailing fields are
   silently ABSENT rather than visibly damaged**. `design_intent` is the *last* of
   the four sections, and its `palette.reference_values` (8 hex slots) is the field
   the composition pipeline and design renderer actually read. Salvaging here would
   hand the site a spec set that is missing precisely the mandatory part, and mark
   the item `complete`. **Failing loudly is better than that.**

3. **The real fix is a smaller unit of work, not a bigger cap.** This one step
   produces four independent documents. Splitting it so each `write_*_spec` step is
   fed by its own bounded generation would make the truncation class
   unrepresentable rather than merely unlikely, and would let a single section be
   regenerated without redoing the classification. Architecture-scope; not attempted
   here.

4. **A floor check on caps.** No step that emits a whole spec document should sit
   below the 8000 mode. An offline lint over `agent_definitions` would have caught
   this in April. Cheap, and it is the check that generalises.

   > **[DONE 2026-08-06 — shipped as a DISTRIBUTION check, not a static floor.]**
   > `scheduled_tasks` row **`fleet-step-token-pressure`**: live, CTE-only
   > (`fire_message=false` — no Kafka, no orchestration, no LLM, no credits), every
   > 6h, delivering to `doc_notes`. It is FIX-058 (`council-seat-token-pressure`)
   > generalised past `review_%`; the two tasks now partition the fleet exactly.
   > Lane: `docs024_key_docs_latest/bugfix_183_step_token_pressure/` (plan, runbook,
   > notes, seed SQL). Register entry **LCO-007**.
   >
   > **Why distribution and not the floor this file asked for.** A static floor
   > needs a judgement about which steps "emit a whole document", and it cannot see
   > a step that outgrows a perfectly reasonable cap — which is what happened here.
   > Headroom subsumes it: a step near its cap is flagged whatever the number is,
   > and a step comfortably inside a 4000 cap is left alone. The floor idea survives
   > in a narrower and sharper form as `bugs_open/205` candidate 1 — steps that
   > configure NO cap at all and silently inherit the transport's hardcoded 2048.
   >
   > **It flags this bug's own case, tested by pinning the window** — as-of
   > 2026-08-02 18:00 it puts `classify_and_extract@6000` at the top (T, n=21, 5
   > truncations); as-of 2026-08-01, BEFORE the first truncation, it already flags
   > it P (n=15, p95 90.0%). So it is a leading indicator on this very bug, not
   > only a post-mortem. Live today the step is correctly ABSENT (cap 32000, last
   > run at 13% of it). Negative control holds: `stage_implement@32000` (p95 24%)
   > does not flag on pressure.
   >
   > **Its first run found a live defect nobody had filed** — `bugs_open/205`:
   > `vet-practice-verifier/extract_and_reconcile` truncating 100% of calls for 34
   > hours against a cap **no one ever configured**, with two poisoned records
   > re-dispatched every few minutes. That is the disconfirmable evidence that the
   > check earns its place.

## How to verify

```sql
-- 1. the cap is live and unshadowed (query above), and
-- 2. the step stops truncating:
SELECT created_at, max_tokens, output_tokens, success,
       left(coalesce(error_message,'-'),80) AS err
  FROM llm_call_log WHERE step_name='classify_and_extract'
 ORDER BY created_at DESC LIMIT 10;
```

**Count truncations from `error_message`, never from `output_tokens >= max_tokens`** —
a truncated first attempt logs `output_tokens = NULL`, so the token comparison
cannot ever match one (`llm_call_log` trap, `MEMORY`; measured fleet-wide 4 vs a
true 94).

## Landmine this leaves behind

A work item that exhausted `max_attempts` against a **config** fault looks exactly
like one that failed for a site-specific reason. `mortgagecalculator.co.uk` sat
`failed 3/3` while `lendzy.co.uk` — same step, same day, same cap — read `complete`
because its attempt 2 happened to land 398 tokens under the line. **Before
concluding a site's classification failed on something about the site, group the
step's failures across sites**; a config ceiling shows up as several unrelated
domains failing the same step in the same window.

## Owner-ruling compliance (2026-07-31)

This file asserts a cross-cutting root cause and was **not** put through the `090`
diagnosis loop. Stated substitution, per the ruling's named escape hatch: the
platform's own error text names the mechanism verbatim (`stop_reason=max_tokens`
with the byte count); the population is small enough to measure **whole** (all 54
calls, not a sample); the competing structural hypothesis was tested against that
population and **refuted**; and the causal chain contains no unread hop — the cap
is one JSON path, read directly, and the shadowing interaction that could have
falsified the fix was checked explicitly. What remains genuinely unknown is marked
`[UNEXPLAINED]` above rather than argued.
