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
