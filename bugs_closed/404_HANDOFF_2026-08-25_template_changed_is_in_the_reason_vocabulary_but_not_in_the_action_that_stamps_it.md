# 404 — `template_changed` joined the re-render reason vocabulary on 2026-08-18 and `create_rerender_items` was never told: route a template fix through it and you get ASSEMBLE-ONLY items that complete green and ship nothing

Filed 2026-08-25 by the `bugs_open/384` lane, found while answering a council objection
(`7553c120`, `prior_art_librarian`) that asked why migration `615` hand-rolled a fan-out
instead of reusing the estate's existing per-page re-render item creator. **The objection was
the right question and the answer turned out to be a defect.**

Not a duplicate of `bugs_closed/024` ("tool template fixes never render to the page"). 024 is
the same family and it is CLOSED — it predates the `template_changed` reason by a month. This
is the gap that OPENED when migration `460` widened the vocabulary on 2026-08-18 without
updating the Go action that stamps reasons onto items.

## The mechanism

`create_rerender_items_action.go` is the estate's generic "file one `page_rerender` per page"
action. When a re-render is caused by a specific component changing, it does two things — and
it gates BOTH on a hard-coded reason list:

```go
// create_rerender_items_action.go:~223
scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""
stampReason := scoped || reason == "cta_links_stale"

keyReason := ""
if stampReason {
    keyReason = reason          // …otherwise the item is ASSEMBLE-ONLY
}
```

`template_changed` is in **none** of those three branches. `[MEASURED 2026-08-25]` the string
`template_changed` appears **0 times** in that file.

So a caller passing `reason=template_changed` + `component_id` gets:

- `scoped = false` → **no component scoping**: items for every page on the site, not the pages
  that actually carry the changed component;
- `stampReason = false` → `keyReason = ""` → **the spec carries no reason** → `page-rerender`'s
  `check_rerender_mode` takes the ELSE branch → **assemble-only**, which re-ships the stored
  `rendered_html` verbatim.

The template change never reaches a single served page. Every item completes. The status is
green. Nothing logs anything.

**That is `bugs_open/283` §13's finding reproduced exactly**, one seam along — 283 measured it
end-to-end (111 items completed, page DEPLOYED, served bytes unchanged) and migration `460` was
written to fix it, by teaching `page-rerender` the new reason AND rewriting
`component-template-fixer.create_rerender` to emit it. **460 fixed the two places it touched
and left this third one carrying the old list.**

## Why nobody has hit it yet

Because the only live producer of `template_changed` is `component-template-fixer`, which files
its items with **raw SQL in its workflow config**, not through this action. `[MEASURED
2026-08-25]` 334 of 338 live `template_changed` items are keyless, which is that raw INSERT's
signature. So the action's stale list is latent — and it is exactly the trap waiting for the
next author who does the reasonable thing and reuses the shared creator.

This lane nearly was that author. Migration `615` hand-wrote its own INSERT; the council asked
why; reading the action to answer showed that reusing it would have shipped 40 assemble-only
items and no visible change. **The hand-rolled version was right for the wrong reason** — it
was written because nothing callable existed from SQL, not because the callable thing was
broken.

## Evidence

- `grep -c template_changed platform/orchestration/actions/create_rerender_items_action.go` → **0**
- Migration `460_template_changed_rerender_reason.sql`, added `03ac7bfea` (2026-08-18), adds the
  reason to `page-rerender.check_rerender_mode` and rewrites the FIXER's query — it touches no Go.
- `create_rerender_items_action.go:~223` — the three-branch gate quoted above.
- Live shape: 334 of 338 `template_changed` `site_work_items` carry `item_key IS NULL`, and only
  4 of 272 in the last 7 days carry a `page_id` `[MEASURED 2026-08-25]`.

## A second, narrower defect in the same area

`component-template-fixer.create_rerender`'s LIVE query (read from `agent_definitions`, not from
`460`'s text — the live row has since gained `p.rebuild_policy IS DISTINCT FROM 'owned'`) has
**no page-status filter at all**. `[MEASURED 2026-08-25]` of 60 live `tool-cta` instances, **16
sit on ARCHIVED pages**; that query would file re-renders for all 16, re-rendering and
re-publishing retired pages and making a retraction self-undoing. That is `bugs_open/098`
exactly, in a place 098's sweep did not reach. Migration `615` carries the filter `615` needed;
the fixer still does not.

## Fix candidates, ordered by what closes the door

1. **Make the reason list one definition, not three.** `page-rerender`'s
   `check_rerender_mode` condition (DB config), `create_rerender_items`' `scoped`/`stampReason`
   gate (Go), and the fixer's raw INSERT (DB config) all encode "which reasons mean re-resolve".
   They are already out of step. A shared exported list in Go plus a migration that derives the
   condition from it makes the next vocabulary addition a one-place change. This is the only
   candidate that makes the bad state unrepresentable.
2. **Add `template_changed` to both branches in the Go action** (`scoped` and `stampReason`).
   One line each, closes the live hazard, leaves the three-copy structure intact — so the
   FOURTH reason repeats this.
3. **Add the page-status filter to the fixer's query** (a migration; `p.status='active'`, or
   `datahelpers.PageWantedLivePredicateFor`). Independent of 1 and 2, and independently correct.
4. Leave it and rely on nobody reusing the action for template fixes. Rejected as a
   recommendation: "operators must remember X" is the failure mode this estate has recorded most.

## How to verify a fix

Do NOT verify at the item. An assemble-only item and a re-resolving item are indistinguishable
by status — both complete. The discriminating checks:

- the item's `spec->>'reason'` is present (assemble-only items carry no reason), and its
  `item_key` ends `_template_changed` rather than `_assemble`;
- the run's `current_step` path reaches `rerender_page_sections`, not the assemble branch;
- **and at the artefact**: `page_components.rendered_html LIKE '%<a literal only the new
  template emits>%'` on a page carrying the component, which is FALSE before and TRUE after.
  Migration `615` + `tool-cta` is a worked example — 15 of 40 pages flipped that predicate on
  2026-08-25 and the count is re-runnable.

## Related

- `bugs_closed/024` — same family, closed, predates the reason.
- `bugs_open/283` §13 — the measurement that motivated `460`.
- `bugs_open/098` — the archived-page resurrection the fixer's query is still open to.
- `LANDMINES.md`, "A template edited by SQL ships NOTHING…" — the prospective form of this.
- Register `REB-002` — the reason gate itself.
- Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/`.


---

# CORRECTION + STRENGTHENING 2026-08-26 — the drift is TWO values wide, and the direction of failure is the point

Two additions, one of them a correction to this file's own headline.

## 1. I understated it: `create_rerender_items` is missing TWO of the five reasons, not one

The live `page-rerender.check_rerender_mode` condition, read from `agent_definitions`
`[MEASURED 2026-08-26]`, is an explicit allow-list of **five**:

```
input_data.spec.reason == 'image_landed'
  OR … == 'section_data_resolved'
  OR … == 'cta_links_stale'
  OR … == 'template_changed'
  OR … == 'literal_markdown'
    → then_step: rerender_sections
    → else_step: render_page          ← assemble
```

Counting each of those five in `create_rerender_items_action.go` `[MEASURED 2026-08-26]`:

| reason | in the live gate | known to the Go action |
|---|---|---|
| `image_landed` | yes | **1** |
| `section_data_resolved` | yes | **1** |
| `cta_links_stale` | yes | **1** |
| `template_changed` | yes (migration `460`, 2026-08-18) | **0** |
| `literal_markdown` | yes (migration `473`, 2026-08-18) | **0** |

So this file was filed naming one missing value and there are **two**, both added on the same
day by different lanes, neither of which touched the Go reader. That makes the "next vocabulary
addition repeats this" argument in fix candidate 2 not a prediction but an observation: it has
already happened twice, in parallel, within one day.

## 2. The generalisation, and it is the reason this class is dangerous rather than merely untidy

Offered by the `bugs_open/384` filing lane (`agentchassis-51`) and **verified here at both
readers before being written down**:

> A mode/reason vocabulary with more than one reader will drift — and **every reader that does
> not know a value fails toward `assemble`, which is the silent direction.**

The asymmetry is what matters. Checked at each reader:

- **Reader 1, the live gate:** an allow-list with `else_step: render_page`. An unknown reason
  takes the else branch → assemble.
- **Reader 2, `create_rerender_items`:** an unknown reason gives `scoped=false` and
  `stampReason=false`, so `keyReason` stays `""` → the item carries no reason at all → assemble.

Both fail the same way, and that way is the one that **completes successfully while changing
nothing**. A vocabulary whose readers failed toward *re-resolve* would be self-announcing: you
would get too many re-renders and notice. Failing toward assemble means the estate's own
preferred, safe, cheap mode is also its silent-failure mode, so drift is invisible by
construction — which is why `bugs_open/283` had to measure it end-to-end (111 items completed,
page DEPLOYED, served bytes unchanged) before anyone believed it.

**Three instances in three days**, all rooted in one value's meaning living in more than one
place with only some copies updated:

1. the original `384` listing case — assemble re-affirming a stale array;
2. this one — `template_changed`/`literal_markdown` unknown to the item creator;
3. `rebuild_blog_listing` hand-writing `"image": ""` beside `queryresolve`'s projection.

⚠ **The third shares the ROOT but not the ASYMMETRY** — it is a hand-copied field value, not a
reason-vocabulary reader, so it has no fail-toward-assemble direction. Grouping all three under
one mechanism is tempting and would overstate the case; they share "one definition, several
copies, some stale", which is enough.

## 3. New fix candidate 0 — the parity test, which is cheaper than all of them

Ahead of the candidates above, because it is small, needs no design decision, and would have
caught both missing values on the day they were added:

**A test that reads the live gate's reason list and asserts the Go reader knows every value in
it.** The estate already has this exact pattern working elsewhere —
`cmd/config-key-audit/optional_budget_cron_parity_test.go` pins a CronJob's literal against the
Go it must match, and CLAUDE.md tells authors to run it for precisely this class of drift. The
same shape here: enumerate the condition's reasons, assert each appears in the action's gate,
fail naming the missing one. It does not fix the three-copy structure (candidate 1 still does
that) but it makes the next divergence loud at commit time instead of silent in production.

**Its own trap, stated so whoever builds it does not walk into this lane's:** the assertion must
read the LIVE condition, not a copy of it pasted into the test. A parity test whose two sides are
both maintained by the same author, in the same file, cannot come out the other way — the failure
this lane recorded twice in `WRONG_CALLS.md` on 2026-08-25.


---

# CORRECTION 2026-08-26b — the 471-item figure is USAGE OF THE WORKING PATH, not blast radius. Zero items reached the stale reader.

The `384` filing lane measured item counts by reason and found **471 items across up to 26
sites carrying a reason the Go reader does not know** (`template_changed` 452, `literal_markdown`
19) — `template_changed` being the estate's second-busiest reason since it landed. They stated
plainly that they had counted items CARRYING the reason, not items that demonstrably shipped
nothing, and that the inference to assemble-only was **mine**, from reading the gate.

**The inference was wrong, and checking it is what this section records.** Census of who created
those items `[MEASURED 2026-08-26]`:

| reason | created_by | items |
|---|---|---|
| `template_changed` | `component-template-fixer` | 383 |
| `template_changed` | hand-run lanes/migrations (`bugfix_384_toolcta_fanout` 40, `bugfix_398_cta_bg_hero_fanout` 9, `news_editorial_features-lane` 4, `bugs_open/383` 4, `283-*` 5, `migration-*-manual` 6, `aiao-contact-propagate` 1) | 69 |
| `literal_markdown` | `generic` 15, `quality-discovery-agent` 4 | 19 |

**Not one was created through `create_rerender_items`.** Every producer above stamps
`spec.reason` directly in its own INSERT, so `check_rerender_mode` sees the reason and routes to
`rerender_sections` correctly. The 471 items are evidence that the reason is heavily USED, via
the path that works — not evidence of anything shipped silently.

The control that makes this legible: `rerender-pages` (which dispatches `create_rerender_items`)
has created **6,428** `page_rerender` items and only **3** carry a reason at all. That is the
action doing its normal job correctly — site-wide refreshes are SUPPOSED to be assemble-only.
The action is not broken in its ordinary use; it is broken only for a caller that passes it a
reason it does not know, and **no such caller exists today**.

**So this bug's exposure is LATENT, exactly as first filed.** What the 471 figure legitimately
supports is the URGENCY argument, not a damage claim: a reason this heavily used is one a future
author is likely to route through the shared creator, which is precisely the reasonable move that
would break silently. Quote it that way or not at all.

**The specific worry that prompted the census is also clear.** dartsonline carries 8
`template_changed` items (not 5), all `complete`, all created by `component-template-fixer` or
`news_editorial_features-lane`, and **all carrying `spec.reason`** — so all routed to
`rerender_sections`. Nothing on that site reported success and shipped nothing by this mechanism.

## ⚠ But the date anomaly they flagged IS a real, bounded finding — a different mechanism

`literal_markdown`'s first items predate migration `473`, which is what taught the gate that
reason (both `473` and `460` landed 2026-08-18). `[MEASURED 2026-08-26]` **7 of the 19
`literal_markdown` items were created BEFORE 2026-08-18**, the earliest on 08-09.

Those 7 carried a reason the LIVE GATE did not yet know, so they took `else_step: render_page`
— assemble — and completed green. That is genuine silent failure, historical and bounded, and
it is **not** this bug's mechanism: here the gate knows the value and the Go reader does not;
there the gate itself did not know it yet. Same failure direction, one layer up, and it is the
only instance in this family where items are known to have taken the silent branch.

Worth someone establishing whether those 7 pages were later repaired by another route before
treating it as outstanding damage. Not chased here.

## What this changes in the fix candidates

Nothing in their ordering. Candidate 0 (the parity test) is if anything better justified: the
defect is latent precisely because every current producer bypasses the shared action, and the
first author who stops bypassing it is the one who gets bitten. A test is how they find out at
commit time rather than in production.


---

# CORRECTION 2026-08-26c — my own control was a LIVE-WINDOW undercount. The conclusion strengthens; the numbers were wrong and had already been relayed.

The control in the section above — *"`rerender-pages` has created **6,428** `page_rerender`
items of which **3** carry a reason at all"* — was measured against `site_work_items` only.
**`site_work_items` is a rolling window: closing a row archives it into
`site_work_items_archive`, out of the table I queried.** So I published the live slice as if it
were the population, in the very section arguing that a count must cover the population it
claims to describe.

Re-run across both tables `[MEASURED 2026-08-26]`:

| source | items | carrying a reason |
|---|---|---|
| `site_work_items` (live) | 6,428 | 3 |
| `site_work_items_archive` | 10,857 | **200** |
| **total** | **17,285** | **203** |

So the honest control is **203 of 17,285**, not 3 of 6,428 — a 2.7× larger denominator and a
67× larger numerator.

## The conclusion does not just survive, it gets stronger

The question this control exists to answer is "did an unknown reason ever reach the stale
reader?". Asked across live AND archive:

- All **203** reason-bearing items from that path carry **`section_data_resolved`** — a reason
  the Go reader KNOWS (first 2026-07-02, last 2026-08-09).
- `template_changed` or `literal_markdown` via `rerender-pages` / `create_rerender_items`:
  **zero rows, live and archive.**

So "no caller routes an unknown reason through the stale reader" now holds over the **full
recorded history**, not merely over the live window — which is a stronger claim than the one I
made, arrived at by correcting the evidence for it. 404's exposure remains LATENT.

## Why this is worth a whole section rather than an edited number

Because the wrong figure had already left this file. The `bugs_open/410` lane picked it up,
carried it into their fix's rationale, and — to their credit — wrote to say they had NOT re-run
it and would attribute it as *supplied by the 384 lane, not independently re-measured*. It was
that offer to re-run rather than relay which sent me back to the query, where the archive was
waiting.

**Two lanes were one relay away from citing a live-window slice as a fleet-wide population.**
The mechanism is in MEMORY as *"a closer census cannot see what it SUCCEEDED at"* and I walked
into it anyway, on a table I had already queried four times that day.

**The check, and it is one line:** any `site_work_items` count that is meant to describe a
POPULATION rather than the current queue must be `UNION ALL`'d with `site_work_items_archive`.
If the number is going into a document, a bug file or another lane's rationale, the archive is
not optional — the live table answers "what is open now", which is almost never the question a
citation is asking.

The query, so the next reader re-runs rather than relays — ⚠ **but re-running is not enough, and
that was proven within the hour.** The `bugs_open/410` lane re-ran the ORIGINAL query
independently, got 6,428/3 to the digit, and recorded it as first-hand confirmation: they
verified the number by making the same population error, and the exact agreement made it more
convincing rather than less. **Re-derive the POPULATION — which tables, which window — not only
the arithmetic over someone else's choice of table** (`WRONG_CALLS.md` 2026-08-26c/d):

```sql
SELECT 'live' AS src, count(*) AS items, count(*) FILTER (WHERE spec ? 'reason') AS with_reason
  FROM site_work_items
 WHERE item_type='page_rerender' AND created_by IN ('rerender-pages','create_rerender_items')
UNION ALL
SELECT 'archive', count(*), count(*) FILTER (WHERE spec ? 'reason')
  FROM site_work_items_archive
 WHERE item_type='page_rerender' AND created_by IN ('rerender-pages','create_rerender_items');
```

---

# CONTRIB 2026-08-26 (`bugfix_404_rerender_reason_vocabulary` lane, taking this bug) — the vocabulary is SIXTEEN values, and the gate-side instance you bounded at 7 is **129**

Taken by a lane opened today; `who-owns.py 404` showed no active workstream. Two of this file's
own conclusions hold under re-measurement and one is much larger than stated.

## 1. HOLDS: the Go-reader arm is LATENT, exactly as your §26b/§26c corrected it to be

Re-derived independently over `site_work_items` UNION `site_work_items_archive`
`[MEASURED 2026-08-26]`: of the reason-bearing `page_rerender` items created via
`rerender-pages`/`create_rerender_items`, **all carry `section_data_resolved`** — which the Go
reader knows. Zero `template_changed`, zero `literal_markdown`. Your correction stands, and so
does the conclusion drawn from it: no caller routes an unknown reason through the stale Go
reader, so **that** arm is a trap for the next author rather than live damage.

## 2. MUCH LARGER: the same asymmetry is ALREADY REALISED at the OTHER reader

§26b bounds the gate-side instance at *"7 of the 19 `literal_markdown` items… historical and
bounded… not chased here"*. Chased:

**`[MEASURED 2026-08-26]` 129 `page_rerender` items carry a `spec.reason` the LIVE GATE does not
know. All 129 were handled by `page-rerender` — the gate's own agent — and 96 COMPLETED.** By the
gate's own structure (`else_step: render_page`) every one took the assemble branch.

Eleven distinct values, and **two of them first appeared in the last two days**, so this is
ongoing rather than historical:

```
verbatim_adoption_deploy        86   light_palette_chrome_replaced  13  (first 2026-08-25)
"migration 415 repointed …"     11   meta_description_corrected      4
"the 20:2x rewrite deployed …"   4   "bugs_open/238: the £149 …"     4
legal_page_publish               3   listing_stale                   1  (first 2026-08-24)
m2_rebuild_safety_proof          1   claims_corrected                1
"section_edit a007f0ff complete + tool-list removed"                  1
```

⚠ **AND WHAT I COULD NOT ESTABLISH, because the distinction matters as much as the number.**
"Took the assemble branch" is measured. **"Therefore shipped nothing" is NOT**, and I am not
claiming it. I tried: the `migration 415` cohort is the best candidate because its own reason
text says *"this page still serves the raw rule"*. Result — 1 of 3 components on **all 11** pages
now carries `--color-primary-ink`, **including the pages whose items were CANCELLED**. Control
and treatment agree, so the marker arrived by another route and the cohort discriminates nothing.
A cohort is only evidence if a marker can be ATTRIBUTED to it, and reaching for the next cohort
until one agrees is how this estate's worst measurements get made.

So the honest statement is: **129 items were routed to the silent branch; whether any of them
needed the other branch is unestablished.** Some of these plausibly WANT assemble. The point is
that nobody can tell — which is the same complaint this file makes about the Go reader.

## 3. THE STRUCTURAL FINDING — `spec.reason` is TWO FIELDS WEARING ONE NAME

Four of the sixteen observed values are **free prose**: whole sentences, a `£` sign, a bug
reference, an operator's note to themselves. Humans are using `reason` as an ANNOTATION while the
gate uses it as a ROUTING KEY. That is why the vocabulary drifts faster than anyone notices — the
field has no closed set to drift *from*.

Three consequences for the fix candidates, offered as design input rather than as a rewrite of
your ordering:

1. **Candidate 0 (the parity test) is necessary and not sufficient.** Keeping Go and the gate in
   step over "the five" leaves the sixteenth free-text value silently assembling for ever.
2. **Candidate 1 (one definition) should also answer "what happens to a reason nobody
   declared?"** Given the fail-toward-assemble asymmetry this file names so well, the safe answer
   is probably not "assemble silently" — **an unknown routing key that completes green is this
   bug in one sentence.** Loud-but-still-assemble would close the observability half without
   changing any routing behaviour.
3. **The vocabulary spans at least THREE item types**, which widens where a definition has to
   reach: `template_changed` also appears on **65 `section_edit`** items, and `literal_markdown`
   appears ONLY on `item_type='literal_markdown'` items — **never on a `page_rerender` item**. So
   the gate's fifth value may not be exercised through this path at all; worth checking before
   asserting the five are symmetric.

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_404_rerender_reason_vocabulary/`.


---

# CLOSED 2026-09-04 — FIXED AND LIVE, verified at the artefact. Moved to `bugs_closed/`.

Closed by the `bugfix_404_rerender_reason_vocabulary` lane. Council **APPROVED** at round 4
(artifact `e1abb1bc-2713-4fda-84b6-f9b85b36129f`, corr `f2e4ac2a-2bfc-4c82-ac99-d5fd7616edef`,
orch `40639f27-fdca-4059-92bd-1a01d9f55f57`, 2026-09-02 16:33:30.187Z — *"approved with 3 advisory
objection(s) — none high-severity"*, 2 abstained). Four rounds, one trail; **every gating
objection in all four was about submission accuracy, none about the design.**

## What the closure rests on — all `[MEASURED 2026-09-04]`, at the artefact, not at git or a tag

| | evidence |
|---|---|
| the Go reader is LIVE | both `agent-chassis` pods (`v1.0.1360`) contain the change's own literal (`sections-rerender vocabulary`), with a **positive control** (a pre-existing literal in the same file: present) and a **negative control** (a nonsense string: absent) in the same `exec`. ⚠ The `build provenance` line has scrolled out of `--tail=3000` on both, so an empty grep there means "not in range", not "unstamped" |
| migration 656 is LIVE at the object | the fixer's `create_rerender` query carries `p.status = 'active'` exactly **once** |
| the live gate is unchanged | five values, byte-identical to `livespec.CheckRerenderModeConditionClause()` |
| the drift guard actually runs | `live-declaration-drift-check` (`0 7 * * *` UTC) at 07:00:10Z: *"probed 16 live object(s) (4 constraint, 2 scheduled_task, 1 trigger_bindings, 2 trigger_fn, 7 workflow); 0 finding(s)"* — and the tree holds exactly **16** Declarations, **7** of them `workflow`, so all three of this bug's declarations are inside the probed set. `compareAllDeclarations` iterates every Declaration regardless of `Phase` and **exits 2** on NO ROWS or NULL, so a clean run cannot mean "could not look" |
| tests | `platform/livespec`'s 14 rerender/reason tests and the actions-package parity tests all pass |

**Fix candidates 0, 1, 2 and 3 all shipped** (commit `ef4236b4d`, migration 656): one definition
in `platform/livespec/rerender_reasons.go` carrying per value whether it scopes by component and
whether it stamps without one; all four readers naming constants, so retiring a value breaks
compilation everywhere rather than silently disarming one gate; three livespec Declarations
(including a **paired count**, because the gate's clause has no terminator and only a count sees
ADDITION); a corpus lint with positive controls naming migrations 460 and 473; and the fixer's
missing page-status filter.

## ⚠ One correction to a shipped artefact that cannot be edited

`editquality` [medium] was right: 656's own comment reads *"[MEASURED 2026-08-26]
component-template-fixer has exactly 1 active row, as do page-rerender and
**availability-discovery-agent**"* while the verification query shown to the council checked
`('component-template-fixer','page-rerender','**rerender-pages**')` — a `[MEASURED]` marker over a
name no shown measurement covered. **Re-measured 2026-09-04 under the migration's own predicate:
all four types have exactly 1 active row**, so nothing false shipped and the guard (scoped to
`component-template-fixer` alone) never depended on it. **The file cannot be corrected — 656 is
applied, and an applied migration is append-only history whose checksum is in
`schema_migrations`.** The correction lives in the lane NOTES.

## Residuals — none of them this bug's mechanism, each with a home

- **An unknown routing key still completes green.** That is `bugs_open/440` / `RFC_062` phase 3,
  spun out by owner decision 2026-09-02. Its migrations `741`/`742` were BUILT, council-APPROVED
  r1, and held on owner ruling D2 (*the 404 lane co-signs*) — **the co-sign was GIVEN 2026-09-04**,
  conditional on one added Declaration: 741's applier step (c) prescribes a `FragmentMatch` for
  `check_routing_key_known`, which is blind to ADDITION for exactly the reason its own step (b)
  exists. Mutation-proved with both controls (a value removed live → 1 finding, so the guard is
  armed; a sixth value appended live → **0 findings, silent**). Remedy: a paired `CountEqual`,
  `ExpectCount` derived from `CheckRoutingKnownConditionClause()` — **7**, because that clause
  carries `== null` and `== ''` besides the five vocabulary values.
- **The WARN has no durable consumer**, and per the 440 lane has fired **zero** times in
  production because every live producer bypasses the Go creator (`bug_historian` [medium]). Not a
  defect in the warning — it is the write-door placement question, and `RFC_062` owns it.
- **The 7 pre-473 `literal_markdown` items and the 129 gate-side items.** Historical, and **no
  discriminating marker exists** — the §CONTRIB 2026-08-26 entry records the cohort that failed to
  discriminate (control and treatment agreed), and reaching for another until one agrees is how
  this estate's worst measurements get made.
- **`spec.reason` is two fields wearing one name** — `RFC_062`, ruled, in flight.

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_404_rerender_reason_vocabulary/`
(NOTES + README only; the NOTES tail is the state).
