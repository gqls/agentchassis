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

  > ⚠ **THIS BULLET IS TRUE ONLY FOR THE THREE ITEM TYPES `loadOpenPageItems` READS**
  > (`needs_page`, `owned_page_review`, `page_build_failed`) — added 2026-09-04, after it
  > misled this lane into a confident, committed, wrong claim. **Every row this bug's damage
  > actually consists of is `unbuilt_internal_link`**, which that function never looks at.
  > That type is governed by `idx_swi_dedup`, whose predicate lists `'unresolved'` among the
  > statuses that **FREE** the dedup slot — and it has its own working drain
  > (`revalidate_unbuilt_link.go`), which closed 20 keys / 76 rows unattended on 2026-09-03.
  > **Before applying this bullet to a population, project `item_type` and check it is one of
  > the three.** Full account: `WRONG_CALLS.md`, 2026-09-03.

Do the census before acting — the split is per item, not per site:
```sql
SELECT s.domain, w.item_key, w.status, left(w.summary,80)
  FROM site_work_items w JOIN sites s ON s.id = w.site_id
 WHERE w.error LIKE '%mechanism-flow%branches%'
    OR w.summary LIKE '[unresolved after%'
 ORDER BY s.domain, w.updated_at DESC;
```

## ✅ CANDIDATE 1 IS FIXED AND LIVE — PROVEN AT THE SERVED ARTEFACT, ON REAL TRAFFIC

**Council verdict: APPROVED** (round 2, 2026-09-03 10:11:56Z, corr
`6de0f6f2-4f37-492a-9cbd-1ae886311a9b`) — *"approved with 4 advisory objection(s) — none
high-severity"*. All four are answered below with censuses rather than assurances.

**The writer is shown the nested shape and produces it.** First proof was one write at
13:24:58Z; **re-censused 14:00Z on real traffic, and the demand control is no longer thin:**

| check `[MEASURED 2026-09-03 14:00Z]` | result |
|---|---|
| mechanism-flow writer calls since the fresh agent pods | **6** (not 1) |
| …carrying the nested exemplar `"branches": [{` | **6 of 6** |
| …still carrying the old flat `"branches": "..."` | **0 of 6** |
| pages built + **deployed** with `branches` stored as an ARRAY | **4**, across **3 sites** |
| …of which any step stored `branches` as a STRING | **0** |
| last actual failure of this defect | **12:23:58Z** — none since |

The four recovered pages, from `page_components.content_data` (steps → branches typing):

| site | page | steps | branches arrays | strings | populated |
|---|---|---|---|---|---|
| cv1.co.uk | `/how-it-works/index.html` | 4 | 4 | 0 | 1 |
| remortgagecalculator.uk | `/what-your-number-means.html` | 5 | 5 | 0 | 2 |
| advertise.co.uk | `/uk-advertising-regulation-map.html` | 6 | 6 | 0 | 3 |
| remortgagecalculator.uk | `/six-month-checklist.html` | 6 | 6 | 0 | 1 |

**Negative control, in the same query:** lendzy.co.uk `/cant-pay.html`, written 2026-09-02
15:46Z, stores **3 steps / 0 arrays / 3 strings** — the pre-fix shape. The query could
therefore have returned the old shape and did not.

**At the served bytes, with a control:** `advertise.co.uk/uk-advertising-regulation-map.html`
— the page this bug named as stuck — returns **HTTP 200, 85,053 bytes**, carrying **7
`branch-label` and 7 `branch-body`** elements. An invented URL on the same domain returns
**404**, so the domain is not 200-ing every path. This page's items were the `failed` shape,
and it re-minted and built **on its own**, exactly as §Unsticking predicted.

**Over-production watch — first answer, and it is the reassuring one.** `[MEASURED
2026-09-03 14:00Z]` **7 of 21** steps across the four pages carry a populated `branches`
array (**33%**). The empty arrays are the omission advice obeyed, not a shortfall. The
accepted risk was that a filled exemplar would make the writer invent decision points; on
this evidence it is being conservative, not over-producing. Re-census as volume grows.

> ⚠ **A trap that nearly cost a wrong verdict, now a landmine.** A failure census keyed on
> `site_work_items.updated_at` showed **3 failures after the fix**. There were none. The
> `error` column **persists** after a row moves on, and `trg_site_work_items_updated_at`
> bumps `updated_at` on **every** write — so old error text resurfaces under a fresh
> timestamp, and two of those three rows were `complete`. **Key a failure census on the
> failure's own event** (`orchestration_states.updated_at` where the error lives), never on
> the work item's `updated_at`.

**This bug stays OPEN** — candidates 2 and 3 are untouched, and §The 52 blocked keys below
now measures exactly how much of the damage candidate 1 cannot reach.

### The four council objections, answered by measurement

| seat | objection (severity) | answer `[MEASURED 2026-09-03 14:00Z]` |
|---|---|---|
| `bug_historian` | single call site of a shared lossy mechanism; other consumers of `item_fields` unaudited (medium) | **Exactly 1** active agent config mentions `item_fields` — `page-content-writer`. The only other Go consumer is `expectedItemFieldsFromComponentSchema` → `reconcileGeneratedItemKeys` (`v3_site_actions.go:8096`, `:2279`), which reconciles key **names** in already-generated content and therefore *wants* the flat projection. No second exemplar-builder exists. |
| `reuse_agent` | census not scoped to per-site **tool forks** (medium) | Forks live in `content_components` as rows, so they were already in scope — proven, not asserted: the census population holds **27 active forks** (and 184 non-forks), and the structured-property census over that whole population still returns **1 row**, `mechanism-flow.steps.branches`, **not** a fork. |
| `guardian` | `agent_definitions` duplicate-active-rows trap not checked before apply (medium) | Settled empirically: migration 724 applied cleanly, and its own `GET DIAGNOSTICS n; RAISE IF n<>1` guard is precisely the check for that trap. `page-content-writer` is not one of the duplicated types. |
| `editquality` | rationale cites `TestStructuredItemShape_NestedObjectProperty`, absent from the sketch (medium) | The test **exists** — `structured_item_shape_test.go:159`, one of 10. The defect was in the submission sketch, not the code. Correctly raised: an unshown test is indistinguishable from an absent one. |

**Residual, accepted, not actioned:** `debug_historian` noted migration 724 edited
`agent_definitions.default_config` with no `snapshot_agent()` backup row, and that its
fail-loud re-run guard means it is not safely re-runnable. Both are true. The migration
carries its own verified rollback and the row was read back after apply; a future editor of
this row should take the snapshot.

## ✅✅ 2026-09-04 16:11Z — FULLY DRAINED. Zero unresolved rows, zero failures, zero human actions.

`[MEASURED 2026-09-04 16:11Z]`, the whole family, every site:

| | |
|---|---|
| `unresolved` rows remaining | **0** (was **251** on 2026-09-03 14:00Z, **175** this morning) |
| `complete` | **275** rows across **56** keys |
| `failed` | 123 rows / 69 keys — terminal and historical; a `failed` row re-mints on its own (§Unsticking) |
| new failures of the defect since the fix | **0** |

**Not one row was touched by a human at any point.** The daily drain ran at 16:08:06Z and
closed all 175 remaining rows `auto:revalidated` in ~3 minutes (loanzy 113, farmerinsurance
63 across 19 keys). Sweep `fba38b50-cb4f-44de-a290-51019c2a0262`, COMPLETED 16:11:05Z.

**This is the decisive refutation of this file's own "52 permanently blocked keys" claim.**
Every one of them recovered unaided. The correction banner below explains the mechanism error;
this is the outcome that settles it.

### ⚠ A near-miss worth more than the result: the sweep was LATE, not dropped

The v1.0.1361 roll restarted the chassis at 16:01:26Z/16:01:53Z, and the sweep was due ~16:06Z
— squarely inside the ~300s window in which CLAUDE.md says a dispatch is **silently dropped**.
At ~16:0xZ a peer session measured, found no orchestration row, and reported the drop; this
lane had primed that reading by flagging the silent-drop risk in advance. It was about to be
escalated to a human for a manual re-trigger.

**The row appeared at 16:08:06Z** — about a minute after the restart window closed, and
*inside* the "absent after ~16:10Z means dropped" threshold this lane had itself supplied. The
tick was deferred by the restart, not swallowed by it.

**The transferable part: a vividly-described failure mode makes an absence look like
confirmation of it**, and it does so for whoever you described it to, not just for you. The
threshold was correct and stated; the reading was simply taken before it. **When you warn a
peer about a silent failure, give them the wait time in the same breath as the symptom** —
and when an absence arrives early, re-check the clock before believing it.

Also checked, because the peer flagged it as explicitly unverified: a dropped dispatch could
NOT have wedged tomorrow's run. `scheduled_tasks` has **no `state` column at all**, so there
is no status field to stick, and the concurrency group has a single member.

## ✅ 2026-09-04 — THE ORIGINAL SYMPTOM IS GONE, AT THE SERVED BYTES, WITH NO INTERVENTION

`[MEASURED 2026-09-04 11:05Z]` **Zero** new failures of this defect since 2026-09-03
12:23:58Z — ~23 hours, spanning the fleet's overnight build traffic.

**The page this bug was filed about is live.** loanzy `/your-rights.html` — active, linked
from live pages, `deployed_at` NULL since 2026-08-18 — **deployed 2026-09-04 04:36:00Z** and
serves **HTTP 200, 116,614 B, with 5 rendered `branch-label` / 5 `branch-body`** elements.
An invented URL on the same domain returns **404**, so this is not a parked-domain artefact.

| site | page | deployed | serves |
|---|---|---|---|
| loanzy.uk | `/your-rights.html` | 2026-09-04 04:36:00Z | 200, 116,614 B, 5 branches |
| loanzy.uk | `/guides/tool-loans-consolidation-guide.html` | 2026-09-04 04:42:50Z | 200, 102,470 B |
| farmerinsurance.uk | `/claims.html` | 2026-09-04 01:18:56Z | 200, 85,055 B |

**Two of the three pages §4 named as stuck are now live, and nobody fired anything at them.**
The third, loanzy `/guides/index.html`, is **not this bug**: it carries **zero
`page_components` rows**, so it never reached the writer. Referred to its owning lane.

### The 175 remaining queue rows are STALE, not blocked — and this closes the correction below

loanzy (113 rows / 14 keys) and farmerinsurance (62 / 18) still read `unresolved`. They are
inert. Every one points at just **two** target pages — the two deployed above — and the daily
drain last ran **2026-09-03 16:06:07Z**, recording `verifier_target_still_unbuilt`. That was
**correct when written**: both targets deployed **9-12 hours later**. The next run (~16:06Z
daily, `review-queue-revalidate-daily`, enabled) should close all 175 as `auto:revalidated`.

Checked, not assumed:
- the sweep drains `status IN ('needs_human_review','unresolved')` — `unresolved` **is**
  included (`workItemRevalidatableStatuses`, `work_items_common.go:143`);
- `max_items` is **1500** in the agent's step config, not the code default of 50; one run on
  2026-09-03 revalidated **440** rows across 15 sites;
- the resolving disjunct is `NOT NeverDeployedPagePredicate` =
  `NOT (deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed')`
  (`datahelpers/links.go:277`) — both targets now fail that predicate on **both** arms.

**Precedent for the prediction:** cv1.co.uk and remortgagecalculator.uk sat in exactly this
state on 2026-09-03, their pages built, and all 20 keys / 76 rows closed themselves
`auto:revalidated` at 16:08Z. **No rows have been touched by this lane, and none should be.**

### What this means for the fix candidates

**Candidate 2 needs restating rather than building.** For `unbuilt_internal_link` a repair
path **exists and works** — `revalidate_unbuilt_link.go`, drained via
`revalidate_review_queue_action.go:324`. The gap this bug asserted ("nothing re-plans or
regenerates") is **not true for this item type**. Whether it holds for others is now an open
question, not a finding. **Candidate 3 is untouched and is the valuable one:** nothing
escalated `/your-rights.html` for the three weeks it sat active, linked and unbuilt — the
recovery, when it came, was incidental rather than triggered.

## ~~THE 52 BLOCKED KEYS~~ — the fix's reach, and the claim this section got WRONG

> ⛔ **CORRECTED 2026-09-03 17:10Z, by the session that wrote it three hours earlier. The
> counts below are right; the word "blocked" is WRONG, and so is everything built on it.**
>
> **These rows block nothing, and 20 of the 52 keys closed themselves while this section
> sat in the repo saying they never would.** By 17:08Z, cv1.co.uk (1 key / 12 rows) and
> remortgagecalculator.uk (19 keys / 64 rows) had **fully drained** —
> `resolution_path='auto:revalidated'`, closed 16:08Z, **no human action**. The estate
> already has the repair path for exactly these items (`revalidate_unbuilt_link.go`, drained
> via `revalidate_review_queue_action.go:324`): once the link's target page builds, it closes
> the item. Candidate 1 made the targets buildable; the drain did the rest.
>
> **The mechanism error, precisely.** §Unsticking's claim that *"an `unresolved` row DOES
> block re-minting"* is correct **for the item types `loadOpenPageItems` governs** —
> `needs_page`, `owned_page_review`, `page_build_failed`. **Every row counted below is
> `unbuilt_internal_link`**, which that function never reads. What governs *it* is
> `idx_swi_dedup`, whose partial predicate lists `'unresolved'` among the statuses that
> **free** the dedup slot: `WHERE item_key IS NOT NULL AND status <> ALL (ARRAY['complete',
> 'verified','rejected','wont_fix','failed','unresolved','cancelled'])`. So for these rows
> the status is explicitly **non**-blocking.
>
> **The cheap check that would have caught it: project `item_type` in the census.** One
> column — all 175 rows are one type, and that type is absent from the three-item list in the
> function being cited. Full account: `WRONG_CALLS.md`, 2026-09-03.
>
> **What is still true:** loanzy.uk (14 keys) and farmerinsurance.uk (18 keys) have not
> drained, and loanzy's three pages remain unbuilt and untouched since August. But the reason
> is **not** a blocking row — it is that nothing has rebuilt their target pages, so the drain
> has nothing to revalidate. **The remedy is a build, not row surgery.** Clearing these rows
> would accomplish nothing; it was proposed to the owner and withdrawn.

**Added 2026-09-03 14:00Z.** The fix stops new occurrences and rebuilds nothing. Until now
that limit was described; here is its size.

`[MEASURED 2026-09-03 14:00Z]` **73 keys** have ever carried this error. **52 of them (71%)
are permanently blocked** by **251 rows in status `unresolved`** — every one branded
`[unresolved after 2 attempts]`, a state deliberately kept in `loadOpenPageItems`' open set
(`reconcile_site_plan_action.go:751-756`) so that it **blocks re-minting** rather than being
re-minted past.

| domain | blocked keys | unresolved rows |
|---|---|---|
| remortgagecalculator.uk | 19 | 64 |
| farmerinsurance.uk | 18 | 62 |
| loanzy.uk | 14 | 113 |
| cv1.co.uk | 1 | 12 |
| **total** | **52 of 73 keys** | **251** |

~~**This does not decay.**~~ **FALSE — and this was the one clause in the paragraph with no
measurement behind it.** It decayed within three hours: 20 of the 52 keys closed themselves.
See the correction banner at the head of this section.

**What replaced it `[MEASURED 2026-09-03 17:08Z]`:**

| domain | keys 14:00Z | keys 17:08Z | how they closed |
|---|---|---|---|
| remortgagecalculator.uk | 19 | **0** | `auto:revalidated`, 16:08Z |
| cv1.co.uk | 1 | **0** | `auto:revalidated`, 16:08Z |
| farmerinsurance.uk | 18 | 18 | — target pages not rebuilt |
| loanzy.uk | 14 | 14 | — target pages not rebuilt |

**So the honest statement of reach is the opposite of the one this section made.** Candidate
1 plus the existing revalidation drain is repairing the backlog unattended, at roughly
20 keys in the first three hours, gated only by whether a site's target pages get rebuilt.

**What loanzy and farmerinsurance actually need is a page BUILD**, which lets the drain do
its work — not row surgery. loanzy's three pages (`/your-rights.html` `needs_rebuild`,
`/guides/index.html` and `/guides/tool-loans-consolidation-guide.html` `planned`, all
`deployed_at NULL`) have not been touched since 2026-08-18/27 and are the worked example:
nothing is retrying them, and no row is stopping it.

⚠ **The demand control has also grown well past what the table above rests on:**
`[MEASURED 2026-09-03 17:08Z]` **29** mechanism-flow writer calls since the fix
(6 + 9 + 11 + 3 across the 13:00–16:00 hours), against the **6** recorded earlier.

---

### How the earlier "it does not work" reading arose (kept, because the mistake is instructive)

> Everything below was measured 12:07–12:20Z and was accurate for those executions. The
> error was mine and structural: **agent work runs in PER-AGENT PODS**
> (`agent-page-build-handler-*`, spawned per job), not in the `agent-chassis` deployment I
> probed, and those executions ran on pods spawned before the roll. The fix was correct and
> deployed the whole time. **Probe the pod that does the work, not the one that shares its
> name** — the estate's own landmine, which I had re-read that morning and still walked into.

## ⛔ SUPERSEDED — the first post-roll observation, and why it read as a failure

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
| farmerinsurance.uk | **21** → **21** | 3 → 3 | 24 |
| remortgagecalculator.uk | **6** → **2** | 13 → 17 | 24 |
| loanzy.uk | **3** → **3** | 6 → 6 | 15 |
| advertise.co.uk | 0 → 0 | 2 → 4 | 4 |
| cv1.co.uk | 0 → 0 | 1 → 1 | 3 |
| mortgagecalculator.co.uk | 0 → 0 | 0 → 0 | 2 |
| leopardessconsulting.co.uk | 0 → 1 | 1 → 0 | 1 |

> **Second figure = `[RE-MEASURED 2026-09-03 14:00Z]`**, three hours after the first, and it
> shows the decay working exactly as the warning below predicted — but only where the site
> got traffic. remortgagecalculator dropped **6 → 2** as siblings aged out and its pages
> rebuilt; farmerinsurance and loanzy did not move at all. advertise rose 2 → 4 *one-away*,
> which is its four **successful** builds landing as `complete` terminal siblings — benign,
> and a reminder that this counter cannot distinguish success from failure.
> **⚠ This is the counter the fix changes the sign of, not the one that matters.** The
> figures that matter are the 52 already-`unresolved` keys above, and those do not decay.

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
- ⚠ **The failure census must be keyed on the FAILURE's own event, not on the work item.**
  `site_work_items.error` persists after the row moves on, and `trg_site_work_items_updated_at`
  bumps `updated_at` on every write, so old error text keeps resurfacing under a fresh
  timestamp (on 2026-09-03 this showed 3 post-fix failures where there were **0**, two of
  them on rows that were `complete`). Count the orchestrations instead:
```sql
SELECT date_trunc('hour', updated_at) AS hr, count(*) AS failures
  FROM orchestration_states
 WHERE error ILIKE '%mechanism-flow%branches%' AND updated_at > now() - interval '24 hours'
 GROUP BY 1 ORDER BY 1;
```
- **The 52 blocked keys — the number that decides whether this bug's damage is over:**
```sql
WITH fam AS (SELECT DISTINCT site_id, item_key FROM site_work_items
              WHERE error ILIKE '%mechanism-flow%branches%')
SELECT s.domain, count(DISTINCT w.item_key) AS blocked_keys, count(*) AS unresolved_rows
  FROM site_work_items w JOIN fam f ON f.site_id=w.site_id AND f.item_key=w.item_key
  JOIN sites s ON s.id=w.site_id
 WHERE w.summary LIKE '[unresolved after%'
   AND w.status NOT IN ('complete','failed','cancelled','rejected')
 GROUP BY 1 ORDER BY 2 DESC;
```
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
