# 454 — the light re-render computes a section plan and drops it, so every page is rendered from its own stored data

**Filed** 2026-09-03 by the `bugs_open/427` session (continuing from
`docs/agent_docs/docs024_key_docs_latest/bugfix_427_event_render/HANDOFF_2026-09-03_continue_here.md`).

**Status: FIXED IN CODE, NOT YET LIVE — stays OPEN.** Fix committed `9831e9ab4`
(2026-09-03). Go changes are inert until an image is rebuilt and rolled; both live
`agent-chassis` builds still carried the defect at 2026-09-03 09:54 UTC. Close it only
when a chassis image containing `9831e9ab4` is running AND a re-render has been shown
to move real resolver data onto a real page.

**Council:** submitted `075cfedd-aef0-4230-b4f1-909ecf68959d`, verdict pending at
filing time (`Council-Submitted:` trailer on the commit; 098 credits it automatically
once approved).

---

## 1. One paragraph

`classifyStoredSection` calls `planSection`, uses the result to decide whether the row
can render, and then throws the plan away. `renderPlannedSection` reads it back out of
the classification struct and gets a zero value. So the re-render's render context is
`base ⊕ stored content_data` with **no freshly resolved data at all**, and the
`content_data` it persists is the stored map unchanged. Every light re-render since
2026-09-02 has rendered each page's own stored data back at itself and reported
success.

## 2. The defect, exactly

`platform/orchestration/actions/rerender_page_sections_action.go`.

`sectionClassification` declares a `plan sectionPlanItem` field. It is **read at
exactly one line in the repository** —

```go
// renderPlannedSection, line ~1501
comp, plan, htmlTemplate := cls.comp, cls.plan, cls.htmlTemplate
```

— and **written at none**. `classifyStoredSection` computes `plan := planSection(...)`,
branches on `plan.Status != "ready"`, then returns having set only `c.comp` and
`c.htmlTemplate`. This is legal Go: a struct field that is only ever read yields its
zero value, and no compiler, vet or lint pass says anything.

Downstream, both readers of `plan.ResolvedData` therefore see `nil`:

```go
// render context: base ⊕ stored content_data ⊕ fresh resolved_data
mergeIntoRenderContext(rc, baseData)
mergeIntoRenderContext(rc, s.contentData)
if plan.ResolvedData != nil {            // never true
    mergeIntoRenderContext(rc, plan.ResolvedData)
}
...
mergedContent := make(map[string]interface{}, len(s.contentData)+len(plan.ResolvedData))
for k, v := range s.contentData { mergedContent[k] = v }
for k, v := range plan.ResolvedData { mergedContent[k] = v }   // ranges over nil
```

**Introduced by** `94f81cc60` (2026-09-02 12:27:53 +0100, "035 P1: extract
classifyStoredSection"), which hoisted `planSection` out of `rerenderFlatSections`'s
loop body into the new function. Before that commit `plan` was an ordinary local of the
loop and reached the render body by scope; after it, the value had to cross a new
struct boundary, and the assignment that would have carried it was not written.

## 3. Why it is invisible

This is the part worth reading, because it is what cost a fortnight.

- **Nothing errors.** No section is carried, so none of the four carry buckets in the
  `resolution` diagnostic (`invalid_template` / `not_found` / `not_ready` /
  `empty_template`) names it — and those buckets exist precisely so that "a run in
  which every section carried" is distinguishable from one that worked (`bugs_open/182`).
  They were the right instrument for the wrong failure: this run does not carry.
- **Every count is exactly what a healthy run reports.** `escalated=false`, `skipped`
  absent, `rerendered` equal to the number of rows on the page.
- **Nothing is blanked.** The stored `content_data` still renders, so the page keeps
  serving its last-good bytes. There is no visible damage, only an absent improvement.
- **So the only observable is a NEGATIVE** — a re-render that changes nothing when the
  resolver's data has changed. You cannot see that in a log; you can only see it by
  knowing what the page SHOULD now say and finding it does not.

Two live mechanisms independently masked it further:

- `planSection`'s own **`carryStored`** (`bugs_open/238`) re-supplies a non-LLM field
  from the page's stored `content_data` when its source resolves to nothing. On a page
  whose fields were already populated, resolved-and-carried and stored-only produce the
  same bytes.
- The **`cta_links_stale`** recompute allocates its own map when it finds
  `plan.ResolvedData == nil`, so CTA repair kept working throughout. The one re-render
  reason anybody was actively watching was the one reason still functioning.

## 4. Population affected

`[MEASURED 2026-09-03]` against `clients_db`, `page_components` with
`build_status='deployed'` joined to `content_components`:

| population | rows | pages | component functions |
|---|---|---|---|
| declares a `query.*`-sourced field | **206** | **196** | **21** |
| declares any non-`llm` source at all | **1,855** | **838** | **82** |

The second row is the true blast radius: `plan.ResolvedData` carries every non-LLM
resolution, not only `query.*` — `site_specs.*` lookups, `renderer`/`static` fallbacks,
`carryStored` carries, and the authoritative hero aliasing (`hero_url` /
`background_image`) that `planSection` writes for any section declaring an image field.

Re-run either census before quoting these numbers; a census goes stale by ADDITION and
reads as current for ever.

```sql
-- the wider one
WITH f AS (
  SELECT pc.id AS pc_id, pc.page_id, cc.function, fld.value->>'source' AS src
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  CROSS JOIN LATERAL jsonb_each(COALESCE(cc.input_schema->'fields','{}'::jsonb)) fld
  WHERE pc.build_status='deployed'
)
SELECT count(DISTINCT pc_id), count(DISTINCT page_id), count(DISTINCT function)
FROM f WHERE src IS NOT NULL AND src <> 'llm' AND src <> '';
```

**Not affected:** `cta_links_stale` (see §3). **Dead rather than wrong since the
regression:** the literal-markdown strip over fresh resolved data (`bugs_closed/184`)
— it only ever ran on `plan.ResolvedData`, so it has been unreachable, though it is
double-gated on a step flag AND `reason=literal_markdown` and so was inert for every
other re-render anyway.

## 5. How it was found

From the far end, not from this file. `bugs_open/427` attached an `event-list`
component (`source: "query.upcoming_events"`) to boxingonline.com's fight-calendar page
with one genuinely qualifying `evidence_base` fact, and dispatched `page-rerender`
three times against a chassis 650 commits fresh. Every run completed with
`escalated=false` and a real `rerendered` count; `content_data` and `rendered_html`
came back byte-identical every time — the old `generic-text-block` keys and 1,813 bytes
of guarded empty state.

That session could not find the mechanism, said so, and named the first thing it had
not tried:

> read `resolveComponent`/`classifyStoredSection` in `rerender_page_sections_action.go`
> line by line rather than by grep, to settle whether the row is being CARRIED (stored
> HTML reused …) rather than freshly rendered

This bug is that read. The answer turned out to be neither of the two options that
session was choosing between — the row was **freshly rendered from stale inputs**,
which is a third state its reasoning did not contain, and which produces exactly the
byte-identical output it had (correctly) said could not discriminate.

**One loose end from that session is now explained.** It reported capturing zero
business-logic log lines from `queryresolve/upcoming_events.go`, whose
`logger.Info("queryresolve: resolved upcoming_events", ...)` fires unconditionally on
every call. Under this root cause the resolver **is** still called (`planSection` runs
in full; only its result is dropped), so that line must have been emitted and the
capture missed it — the pod-log capture was the faulty instrument, not the code. Do not
carry "the query resolver never ran" forward; it is refuted.

## 6. The fix

Commit `9831e9ab4`. One assignment, placed where the same function already sets its two
siblings on the success path:

```go
	c.comp = comp
	c.htmlTemplate = htmlTemplate
	// … the plan is LOAD-BEARING, not bookkeeping …
	c.plan = plan
	return c
```

Placement is deliberate: the three carry branches above return early with a `carryKind`
and `renderPlannedSection` is never reached for them, so the plan only means anything
once classification has decided the row renders.

**Pinned by** `platform/orchestration/actions/rerender_page_sections_resolved_data_test.go`,
two tests at two levels — the classification itself (where the value is lost) and the
rendered bytes plus the persisted `content_data` (where it is user-visible). Asserting
only the first would leave `renderPlannedSection`'s merge unguarded, and that merge is
the half a future refactor is equally free to drop.

The test uses boxingonline.com's real component and a real register fact, so it
reproduces 427's live symptom offline. Its event date is **computed** from `time.Now()`
rather than the literal `2026-10-31` the live register carries — a hardcoded future
date becomes a past date, and the resolver's own `date.Before(today)` rule would then
turn the test into a vacuous pass.

**Mutation-proven at committed HEAD, not in the working tree** (a neighbouring session
had `platform/orchestration/datahelpers/claims.go` dirty and non-compiling at the time,
so `go test` in the tree could say nothing about this change either way):

```bash
# the test alone against unfixed HEAD a49fc3a36 — FAILS all four assertions
scripts/verify-head-builds.sh --with platform/orchestration/actions/rerender_page_sections_resolved_data_test.go \
  --test ./platform/orchestration/actions/
# the test plus the fix — ok, whole tree green
scripts/verify-head-builds.sh --with platform/orchestration/actions/rerender_page_sections_action.go \
  --with platform/orchestration/actions/rerender_page_sections_resolved_data_test.go \
  --test ./platform/orchestration/actions/...
```

## 7. Why this was not put through `090_TRIGGER_needs_diagnosis`

CLAUDE.md's diagnosis-before-debugging section, and the owner ruling of 2026-07-31,
require either a loop run or **a plainly stated reason for substituting equivalent
first-hand verification** before a `bugs_open/` file asserts a cross-cutting root cause.
The reason here, stated rather than omitted:

The claim is not an inference about behaviour, it is a **property of the source text** —
a struct field read at one line and assigned at none, which one grep settles and which a
failing test then demonstrates end to end. There is no hypothesis for the loop to
refute: the disconfirming result would have been a second grep hit, and there is none.
The handoff that reached this session *did* nominate a `090` run for the symptom, and
that was the right call at the time; the read it also nominated arrived at the mechanism
first, and a diagnosis run against a symptom whose cause is already mutation-proven
would be spending credits to re-derive a test that is now committed.

What that substitution does **not** cover, and what a reader should treat as unverified:
whether any code written between 2026-09-02 and today was tested against, or has
silently adapted to, the broken behaviour. I looked at this action's own callers, not at
every consumer of the pages it writes.

## 8. Verifying it once it ships

1. A chassis image carrying `9831e9ab4` is running:
   `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`
   then `git merge-base --is-ancestor 9831e9ab4 <the stamp>`. (The line is a STARTUP
   line and scrolls; if it is out of range, probe the binary with a known-present and a
   known-absent sha as controls.)
2. Dispatch `page-rerender` with `spec.reason=section_data_resolved` at boxingonline.com's
   `tool-fight-calendar` page (recipe in
   `docs/agent_docs/docs024_key_docs_latest/bugfix_427_event_render/RUNBOOK_bugfix_427_event_render.md`).
3. Read the **artefact**, not the status:

```sql
SELECT pc.content_data->'items' AS items, length(pc.rendered_html)
FROM page_components pc
WHERE pc.page_id='4b74ff1f-455a-4bb2-b81d-e1d0ec824f33' AND pc.slot_name='event-list';
```

`items` must be non-empty and `rendered_html` must no longer be 1,813 bytes. Then curl
the served page — a DB row is not what a customer sees.

## 9. Relations

- **`bugs_open/427`** — the symptom this was found through. 427's one remaining open
  defect IS this bug; it cannot close until this ships.
- **`features_open/035`** (the composition-walk decomposition, P1) — the lane that owns
  `rerender_page_sections_action.go` and made the extraction. Not blame: the extraction
  is careful work with an unusually rigorous commit message. See §10 for the part of
  that rigour that could not have caught this.
- **`bugs_open/182`** — the carry-bucket diagnostic. Right instrument, wrong failure;
  this run does not carry, so no bucket names it.
- **`bugs_open/238`** — `carryStored`, one of the two mechanisms that masked this.
- **`bugs_closed/184`** — the resolved-data markdown strip, dead since the regression.

## 10. The transferable pattern (also filed to 016b §9 and LANDMINES)

**Hoisting a computation into a helper moves its RESULT across a struct boundary, and
Go will not tell you if the result stops travelling.** A line-count reconciliation of an
extraction — even a careful one that accounts for every removed line, as
`94f81cc60`'s does — measures the code that MOVED. A value that stopped being carried
has no line of its own to be missing from either side, so it is invisible to exactly the
check the extraction was audited with.

The cheap check, when a refactor introduces a struct to carry state across a new
boundary: **for every field on it, grep for reads and writes separately.** A field with
reads and no writes is the defect, and it takes one command:

```bash
grep -n '\.plan\b' <file>     # one hit: a read. No assignment anywhere.
```

## 11. Council verdict — APPROVED, 2026-09-03, and the three advisories adjudicated

`075cfedd-aef0-4230-b4f1-909ecf68959d`, **APPROVED** at 2026-09-03 10:05:35Z, round 1.
`decided_by: "approved with 1 advisory objection(s) — none high-severity"`; 12 seats
reviewed, 5 abstained, `gated_by_truncation: false`. Twelve of thirteen verdicts were
approve. The three advisories are all about the SUBMISSION's sketch rather than the code, and
none of them changes it — but "no action needed" is a conclusion that has to be shown, so
each is answered here rather than waved past.

**1. `editquality`, MEDIUM — "the rationale promises two tests, the sketch contains only the
second."** Correct about the submission and wrong about the artefact, entirely my fault for
sketching one test where I had written two. The committed file
(`rerender_page_sections_resolved_data_test.go`, commit `9831e9ab4`) carries **both**, and
both are named in the `symbol` field the seat quoted:

- `TestClassifyStoredSection_ReturnsTheResolvedDataItComputed` — the classification-level
  assertion the seat was worried was missing. It fails today with
  *"classifyStoredSection computed a plan and threw its ResolvedData away"*.
- `TestRerenderFlatSections_FreshResolvedDataReachesHTMLAndPersistedRow` — the user-visible
  one, the only body the sketch showed.

The seat's reasoning — that without the classification-level assertion a future refactor could
re-break `classifyStoredSection` while leaving the merge intact — is exactly why the first test
was written, so the objection is right on the merits and wrong only about what shipped. **The
lesson is the runbook's own, and I did not follow it: reviewers judge the SKETCH; it is the only
view of the code they get.** A second sketch block would have cost four lines and saved a
MEDIUM.

**2. `editquality`, LOW — "the sketch assumes `plan` is already a local in scope."** It is:
`plan := planSection(ctx, s.slotName, thisSection, comp, resolver, logger)` is in the same
function body, twenty lines above, and the `plan.Status != "ready"` guard between them is what
the sketch's `@@` header elided. Settled by compilation rather than by argument — the fix and
both tests pass against committed HEAD (`verify-head-builds.sh --test`, HEAD `13aac933f`, "ok …
7.215s").

**3. `reuse_agent`, LOW — "this test belongs in the action's existing test file rather than a
freshly split-off one."** A reasonable default, and measurably not this package's convention.
`[MEASURED 2026-09-03]` there is **no** `rerender_page_sections_action_test.go`. What exists is
one file per concern, and this action already has four:
`rerender_page_sections_base_data_test.go`, `_removed_test.go`, `_resolve_test.go`,
`_scan_completeness_test.go`, plus `rerender_strip_gate_test.go` alongside. A fifth, named for
its concern, follows the established pattern; folding it into a monolith would be the departure.
**No change.**

**4. `bug_historian`, MISSING (not an objection) — "no prior council precedent check possible
without a code_checks/SQL round-trip on `classifyStoredSection`; flagging for a human, since
`bugs_open/044`, `054`, `087` are close cousins by shape."** Fair, and named here so it is not
lost. `044` (on_missing policy drift between the generic and `query.*` paths), `054` (a
query-sourced list resolving empty and being stored as an empty slice) and `087` are all about
`plan_sections` data reaching or failing to reach a page — which is the same neighbourhood,
though none of them is about the plan being computed and discarded. Anyone extending this file
should read them before assuming this was the only way for resolved data to go missing.

**No `Council-Reviewed:` trailer exists on `9831e9ab4`**, and there will not be one: the commit
predates the verdict and carries `Council-Submitted: 075cfedd…`, which the 098 report resolves
to the approval at report time with no amend (forward-only forbids one). That is the designed
path, not an omission.
