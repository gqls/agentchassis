# 437 — mechanism-flow's `steps[].branches` contract is unwritable by its own writer: 119 failed builds on six sites in 14 days, pages stuck `planned` for weeks

**Filed 2026-09-02 by the loanzy lane.** Diagnosis loop NOT run — substituted first-hand
verification, stated per the 2026-07-31 ruling: the error is the system's own, verbatim and
recurring; the spread is a one-query census; the stuck pages are lifecycle rows read directly.
Found investigating why 3 of loanzy's 30 active pages 404 (build_status `planned`/`needs_rebuild`,
deployed_at NULL since 2026-08-18) while live pages link one of them (`/your-rights.html`) inline.

## The chain, verified at the rows
1. `page-build-handler` fails, verbatim error: `failed to execute action render_component:
   component "mechanism-flow": content does not match the declared field type(s) —
   steps[0].branches: declared array (items: object), got string; steps[1].branches: …`
2. The failure repeats until the two-strike arm brands the item
   `[unresolved after 2 attempts]` — the repair queue looks handled; the page stays unbuilt.
3. **Spread [MEASURED 2026-09-02]: 119 `content does not match the declared field type`
   failures in 14 days**, mechanism-flow/branches on six sites: remortgagecalculator.uk 53,
   loanzy.uk 35, farmerinsurance.uk 24, cv1.co.uk 4, mortgagecalculator.co.uk 2,
   advertise.co.uk 1.
4. Symptom at the artefact: ACTIVE pages that never deploy — loanzy `/your-rights.html`
   (linked from at least two live pages' body copy), `/guides/index.html`,
   `/guides/tool-loans-consolidation-guide.html` — the INVERSE of closed 359
   (there: retired-but-serving; here: wanted-but-never-served, and nothing escalates it).

## Mechanism (~~narrowed, not proven to the line~~ **PROVEN TO THE LINE 2026-09-03 — and it is the second branch: the prompt never learned the nested shape**)

> **RESOLVED 2026-09-03 by the `bugs_open/437` session, at the artefact.** This section
> offered two candidates — "the schema changed under the writer" or "the writer's prompt
> never learned the nested shape". **It is the second, and the writer is obedient
> throughout.** The chain, each link read rather than inferred:
>
> 1. `mechanism-flow.input_schema` is CORRECT and has not moved: `steps[].branches` is
>    declared an array of objects `{body,label}` with the description *"a decision point:
>    two or more outcomes, rendered side by side"* (live row, read 2026-09-03).
> 2. `extractArrayItemFields` (`platform/orchestration/actions/plan_sections_action.go:3277`)
>    projects an element schema to a flat `[]string` of NAMES. `llmFieldSpec.ItemFields`
>    carries only names, so a nested type — and every nested description — is dropped, and
>    is **unrepresentable in the return type**.
> 3. page-content-writer's prompt builds its Output Format exemplar from those names, and
>    therefore renders:
>    `"steps": [{ "body": "...", "branches": "...", "marker": "...", "note": "...", "title": "..." }]`
>    — **the prompt itself declares `branches` a string.**
> 4. The evidence, one failing run end to end: `llm_call_log`
>    `34f25815-42d3-4057-b42a-b8b42189ae7e` (page-content-writer, 2026-09-02 19:07:30Z,
>    advertise.co.uk). `prompt_rendered` line 234 is the exemplar above; `response_text`
>    obeys it — `"branches": "Broadcast ads follow the BCAP Code. Almost everything else
>    follows the CAP Code."` The gate at `v3_site_actions.go:2434` then refuses, correctly.
>
> **So the 119 failures measure the INSTRUCTION, not the model.** That is also why the
> count is so high with no lucky passes: a deterministic exemplar produces a deterministic
> reply. Anyone reading §3's census as evidence of an unreliable writer will look in the
> wrong place — as this file's own first candidate did.

`bugs_closed/260` remains the ancestor class (one mistyped LLM field), whose fix made the
render REFUSE loudly instead of degrade silently; this bug was the refusal with no working
repair path behind it. `bugs_open/348` is the adjacent-but-different arm (refusals that
report complete).

## Fix candidates (ordered by what closes the door)
1. ~~Make the bad state unwritable at the WRITER~~ **BUILT 2026-09-03 — AND IT IS LIVE AND NOT
   WORKING. See §POST-ROLL below before doing anything with this candidate.** ~~inert until the next
   chassis roll.~~ Register entry `PBP-052`; council corr
   `6de0f6f2-4f37-492a-9cbd-1ae886311a9b` (submitted alongside the commit).**
   The remedy is narrower than this line guessed: nothing needed to change about
   mechanism-flow, and no coercion was added. `datahelpers.StructuredItemShape` renders the
   nested element shape as a JSON skeleton plus one sentence per structured property
   (carrying its schema description, which the flat projection dropped); `plan_sections`
   carries both on `llm_field_specs` as `omitempty` keys; migration **724** teaches the two
   prompt sites to use them when present. `[MEASURED 2026-09-03]` exactly **1** live
   component qualifies, so every other component's prompt is byte-identical. The two halves
   deploy in either order (absent keys, `{{if}}` guards — proven by test, not argued).
   ⚠ **Coercion was deliberately NOT chosen** and the reason is in `bugs_closed/260` §5
   candidate 3: silently rewriting writer output hides the contract violation. Here there
   is no violation left to hide — the contract now says what it means.
   **How to verify it shipped:** a fresh writer run's `prompt_rendered` shows
   `"branches": [{ "body": "...", "label": "..." }]`, and a previously-failing page builds
   with `branches` stored as an array in `page_components.content_data` — read the artefact,
   not the work-item status.
2. A repair path for type-mismatch refusals: today's only outcome is fail → two-strike →
   unresolved; nothing re-plans the section or regenerates the field with the error in hand.
3. Escalation gap (its own small bug if split): a page ACTIVE + `planned`/`needs_rebuild` +
   never deployed for N days, while other live pages link it, surfaces nowhere.

## Unsticking the six sites — candidate 1 does NOT do it, and the two arms behave differently

**Added 2026-09-03, from reading the producers rather than assuming.** The fix stops new
occurrences; it rebuilds nothing. What happens to the existing stuck pages splits in two:

- **A `failed` row does NOT block re-minting.** `failed` is terminal, so it is excluded
  both from `idx_swi_dedup` (migration 157) and from `loadOpenPageItems`' open set
  (`reconcile_site_plan_action.go:757-763`). `reconcile_site_plan`'s next sweep re-mints
  `needs_page:<name>` on its own, born `triaged`, via a raw INSERT subject to neither the
  attempt ladder nor the two-strike arm. **advertise.co.uk's
  `uk-advertising-regulation-map` is this shape** (both its items are `failed`), so it
  needs no surgery — the sweep will pick it up once the fix is live. There is **no re-arm
  route from `failed` anywhere in the code**, and none is needed: a fresh row is the
  supported path. Never UPDATE the terminal rows.
- **An `unresolved` row DOES block it.** `unresolved` is deliberately kept in the open set
  (`:751-756`), precisely because that raw INSERT has no two-strike of its own. So a page
  branded `[unresolved after 2 attempts]` will sit unbuilt for ever, fix or no fix, until
  someone closes the branding row. **That is the deliberate, separate state-changing step
  this fix does not smuggle in**, and it should follow a verified build on one page rather
  than precede it.

Do the census before acting — the split is per item, not per site:
```sql
SELECT s.domain, w.item_key, w.status, left(w.summary,80)
  FROM site_work_items w JOIN sites s ON s.id = w.site_id
 WHERE w.error LIKE '%mechanism-flow%branches%'
    OR w.summary LIKE '[unresolved after%'
 ORDER BY s.domain, w.updated_at DESC;
```

## ⛔ POST-ROLL 2026-09-03 — THE FIX DID NOT TAKE EFFECT ON FIRST OBSERVATION. CANDIDATE 1 IS NOT CLOSED.

> **⚠ NARROWED 12:55Z, SAME DAY — read this before the section below.** The measurements
> below are sound but were taken 12:07–12:20Z, minutes after the deployment roll, and I had
> **missed that agent work runs in PER-AGENT PODS** (`agent-page-build-handler-*`,
> `agent-page-content-writer-*`, spawned per job) rather than in the `agent-chassis`
> deployment. My binary probe used a deployment pod — the wrong one. Re-probed on a real
> `agent-page-build-handler` pod at 12:55Z: **the fix literal IS present, with control.** So
> those failing executions plausibly ran on agent pods spawned BEFORE the roll, which would
> reconcile every observation below with a fix that is simply correct. Not proven (the pods
> are gone), but it now leads the hypothesis list. No mechanism-flow page has been written
> since the fresh pods came up, so the fix is **un-exercised, not disproven**. Neither agent
> is image-pinned and both rows read `v1.0.1358`. Full account:
> `docs024_key_docs_latest/bugfix_437_writer_prompt_nested_shapes/HANDOFF_2026-09-03_continue_here.md`
> §ADDENDUM.

**Read this before trusting any earlier line in this file about candidate 1.** The chassis
rolled to `v1.0.1358` at 12:06Z. The fix is provably present and the defect is unchanged:
builds are still failing with the identical error, and the writer prompt still carries the
OLD flat exemplar.

**What IS confirmed live (each with a control, so none of these is the gap):**
- Image `v1.0.1358`, revision `d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`.
  `git merge-base --is-ancestor a0044e73b d0252fd4d` → **PASS**; negative control (current
  HEAD) correctly ABSENT.
- **Binary probe on the running pod:** the helper's literal `never a sentence of prose` is
  PRESENT in `/proc/1/exe`; long-lived control present; nonsense control absent. The struct
  tag `value_shape` is present (3×, same count as the long-standing `item_fields`).
- **The built revision's own tree** contains both the helper (`func StructuredItemShape`) and
  **2** references to it in `plan_sections_action.go`, at the single `llmFieldSpecs = append`
  site (line 2723).
- **No service skew:** all three deployments running the chassis image (`agent-chassis`,
  `business-intel`, `vet-intel`) are on `v1.0.1358` at the same revision.
- **Migration 724 is intact in the live row** — nested exemplar 1, item_notes tail 1,
  pre-437 spelling 0, flat arm 1. Not reverted.
- **The helper works on the EXACT live schema bytes.** Dumped `input_schema` from the live
  `mechanism-flow` row, ran it through `SchemaContentFields` → `StructuredItemShape` in a
  throwaway test: returns the correct skeleton
  `[{ "body": "...", "branches": [{ "body": "...", "label": "..." }], … }]` and the correct
  note. So the helper logic is NOT the defect.
- **The live schema is unchanged** (one active `mechanism-flow` row; `branches` still
  `type: array` with `items.properties`; `steps.source = llm`).
- **Timing is not the explanation.** Five post-roll orchestrations (12:07:57 → 12:20:25Z),
  all started AFTER the pods came up. One of them (`29a88d1e`, 12:20:01Z) ran the
  `plan_sections` action itself.

**The observed failure, stated precisely:** in `29a88d1e`'s own `section_plan` output (the
step's real `output_field` — NOT the `plan_sections` key, which is a different thing and
cost me a wrong turn), the **mechanism-flow** section's `steps` spec is:

```json
{"name":"steps","type":"array","required":true,"on_missing":"skip_field",
 "item_fields":["body","branches","marker","note","title"]}
```

`item_fields` is correct and `value_shape` / `item_notes` are **absent**. Since both are
`omitempty`, absence means `StructuredItemShape` returned empty at runtime — on a field
where the same helper, given the same schema out of the same DB row, returns a shape.

**So the contradiction to resolve is exactly this:** the deployed binary contains the call
and the helper; the helper works on that schema; the call site is the only one; and the
emitted spec has no shape. One of those four is false in production and I could not
determine which by inspection.

**Ranked hypotheses for whoever picks this up:**
1. **`comp.InputSchema` at runtime is not the schema I probed with.** The one concrete
   anomaly found: the component payload carried in the plan serialises
   `component.input_schema` as a **JSON STRING**, not an object (`jsonb_typeof` = `string`,
   `? 'fields'` = false). If the loader hands `plan_sections` a differently-shaped schema
   than the raw DB row, `extractArrayItemFields` could still succeed (it only needs
   `items.properties`) while `StructuredItemShape`'s first guard —
   `declaresArray(fieldDef["type"])` — fails. **Start here.** Note this guard is STRICTER
   than `extractArrayItemFields`' entry condition, which is a real asymmetry in my design
   regardless of whether it is the cause.
2. The `section_plan` being read was not produced by that step's execution (carried from a
   parent, or a cached/echoed result).
3. Something between the action's return and serialisation drops the keys.

**The next experiment, because inspection is exhausted:** add a temporary `logger.Warn` in
the `source == "llm"` branch printing `fieldName`, `fmt.Sprintf("%T/%v", fieldDef["type"],
fieldDef["type"])` and whether `fieldDef["items"]` is a map — or run `plan_sections`'
resolver locally against the live DB row. Either answers hypothesis 1 in one run. **Do not
add more queries against `orchestration_states`; that avenue is spent.**

**Status of the two halves:** migration 724 is applied, verified, and HARMLESS while the Go
side emits nothing (the `{{if}}` guards render exactly the old prompt — which is precisely
what post-roll observation confirms). No rollback is needed or advised.

## ⚠ THE RE-MINT WINDOW IS A HAZARD, AND IT IS NOT GATED BY ANYONE'S RESTRAINT (added 2026-09-03)

**Raised by the `portfolio_positioning` lane, measured here.** While candidate 1 sits inert
awaiting a roll, the estate keeps re-minting work for these pages **automatically**, and
each automatic attempt fails on the unfixed writer and burns one sibling toward the sticky
`[unresolved after 2 attempts]` brand. Nobody has to fire anything for this to happen —
advertise.co.uk was re-minted at 2026-09-03 10:34:50Z by a discovery sweep
(`e75f5880`, `needs_content_page`, source **`page-rerender-empty-skip`**), 51 minutes
before this was written.

**Note the route, because the obvious prediction was wrong.** I told that lane
`reconcile_site_plan` would re-mint. The re-mint came from
`fileBuildAskForEmptyPage` (`rerender_single_page_action.go:1276`) instead — a *different*
producer, filing a *different item type* on a *different key shape*
(`needs_content_page:<page-uuid>`, not `needs_page:<name>`). **So "which producer re-mints
this page" is not a safe thing to reason about; several can, and they do not share a key.**

**Why the key shape matters, and it is the whole mechanism:** the two-strike arm counts
terminal siblings **on the same `item_key`, within a rolling 7 days**
(`load_work_item_actions.go:1980-2036`), and brands the incoming row `unresolved` at
`terminalCount >= 2`. An `unresolved` row is then kept in `loadOpenPageItems`' OPEN set
(`reconcile_site_plan_action.go:751-756`), so it blocks re-minting rather than being
re-minted past. A fresh key starts its own ladder from zero — which is why advertise's new
`needs_content_page` is not in immediate danger — but a key that has already failed twice
is one automatic sweep away from being stuck for good. This producer goes through the
shared `writeWorkItem` door, so it **does** inherit the arm.

**How much of the door has already closed `[MEASURED 2026-09-03 ~11:00Z]`** — keys touched
by this defect, by how many terminal siblings they already carry in the rolling window:

| domain | keys AT/PAST the brand threshold (≥2) | keys one away (1) | keys touched |
|---|---|---|---|
| farmerinsurance.uk | **21** | 3 | 24 |
| remortgagecalculator.uk | **6** | 13 | 24 |
| loanzy.uk | **3** | 6 | 15 |
| advertise.co.uk | 0 | 2 | 4 |
| cv1.co.uk | 0 | 1 | 3 |
| mortgagecalculator.co.uk | 0 | 0 | 2 |
| leopardessconsulting.co.uk | 0 | 1 | 1 |

**30 keys across three sites are in the window where the next automatic re-mint is born
branded and stuck.** That is the argument for treating the roll as more urgent than a
normal inert fix: the cost of waiting is not "the pages stay broken", it is "more of them
become unreachable by the automatic recovery that would otherwise have fixed them".

> ⚠ **THIS TABLE HAS A SHORT HALF-LIFE AND MUST BE RE-MEASURED, NOT QUOTED.** The threshold
> counts a **rolling 7-day** window, so a key sitting at 3 terminal siblings today drops
> below 2 within days as they age out, and becomes safely re-mintable again with nobody
> doing anything. The number moves in BOTH directions — down as siblings age out, up on
> every automatic sweep — so it is a snapshot of a moving quantity, not a backlog. Re-run
> the query in §Verify before acting on it. What does NOT decay is an `unresolved` row once
> written: those stay open, and stay blocking, indefinitely.

**Deliberately NOT done here:** no held item types, no closed rows, no touched sites. Every
one of these belongs to another lane, the correct sequence is still fix-then-verify-then-
unstick, and a hold on a shared item type is a config change to a shared seam that wants
its own review. This section exists to make the cost of delay legible, not to license
pre-emptive surgery.

## Verify
- The error: `SELECT left(error,300) FROM site_work_items WHERE error LIKE '%mechanism-flow%branches%' ORDER BY updated_at DESC LIMIT 1;`
- **The fix, at the prompt** (the honest post-roll check — the served page is downstream of
  three more steps): `SELECT prompt_rendered LIKE '%"branches": [{%' FROM llm_call_log
  WHERE agent_type='page-content-writer' AND prompt_rendered LIKE '%mechanism-flow%'
  ORDER BY created_at DESC LIMIT 1;`
- **The re-mint window (the §hazard table above, which decays — re-run, never quote):**
```sql
WITH fam AS (SELECT DISTINCT site_id, item_key FROM site_work_items
              WHERE error ILIKE '%mechanism-flow%branches%')
SELECT s.domain,
       count(*) FILTER (WHERE t.terminal_7d >= 2) AS at_or_past_threshold,
       count(*) FILTER (WHERE t.terminal_7d = 1)  AS one_away,
       count(*) AS keys_touched
  FROM fam f JOIN sites s ON s.id=f.site_id
  JOIN LATERAL (SELECT count(*) FILTER (WHERE w.status IN ('complete','failed')
                                          AND w.created_at > now() - interval '7 days') AS terminal_7d
                  FROM site_work_items w
                 WHERE w.site_id=f.site_id AND w.item_key=f.item_key) t ON true
 GROUP BY 1 ORDER BY 2 DESC;
```
- **The census, with a DEMAND control** — a post-fix zero in the failure census is equally
  consistent with "no mechanism-flow page has been built since", so count the writer runs
  in the same window or the zero proves nothing.
- The spread: the §3 census query (same predicate, GROUP BY site).
- The stuck pages: `SELECT url, build_status, deployed_at FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='loanzy.uk' AND p.status='active' AND p.deployed_at IS NULL;`
