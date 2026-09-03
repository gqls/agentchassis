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

## 12. VERIFIED LIVE, 2026-09-03 — the fix is proven at the artefact, and it stays OPEN for one reason only

**A fresh chassis rolled at 12:18Z and the fix is in it.** Standing pods
`agent-chassis-554857f96f-{kx69c,mdc6d}`, image `v1.0.1358`, both reporting
`git_commit = d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`, and
`git merge-base --is-ancestor 9831e9ab4 d0252fd4d` exits 0.

> ⚠ **The stale-row landmine fired exactly as documented and would have given the wrong
> answer.** A bare newest-first read of `service_binary_capabilities` returned six rows for a
> **spawned** `agent-image-build-handler` pod still carrying the OLD commit `7bf1ff67`. Filter
> to the standing pods by name (`pod_name LIKE 'agent-chassis-<replicaset>-%'`) — the two
> standing pods agree with each other, and that agreement is the signal.

### The proof, from the run's own output rather than from a status

Dispatched `page-rerender` / `reason=section_data_resolved` at the fight-calendar page
(correlation `be75b209-1c52-4563-a7b3-bd00902a0367`). `rerender_sections` reported
`rerendered:2 carried:0 escalated:false` — the same counts a broken run reported for a
fortnight, which is why the counts are not the evidence. **The `sections_metadata` is:**

| slot | content_data keys | items | rendered_html |
|---|---|---|---|
| `event-list` | `content, heading, **items**` | **1** | **2,498** bytes |
| `hero-tool` | +`background_image`, +`hero_url` | — | 3,859 bytes |

Against the pre-dispatch control taken minutes earlier: `event-list` was `content, heading`,
**0** items, **1,813** bytes (`md5 ee2ec06829ff94ad8247945452b95915`), unchanged across three
dispatches yesterday.

**So the defect is closed on both halves of its own claim.** The query-sourced `items` field
populates, and — the part that confirms the blast-radius argument rather than just the symptom
— `hero-tool` simultaneously regained `hero_url`/`background_image`, the authoritative hero
aliasing `planSection` writes for any section declaring an image field. **Two sections, two
different non-`llm` sources, both restored by one line.** That is the 1,855-row population
made visible on the one page that could see it.

### Why this stays OPEN

**The save was refused**, by a guard that went live in the same image:

```
step save_sections failed: failed to execute action save_page_sections:
OWNED_PAGE_GUARD: page tool-fight-calendar is page_type=tool with no tool component:
a generic section save would publish prose about a tool that is not there.
```

That is `bugs_open/450`'s `pageRefusesGenericBuild` (`owned_page_guard.go:161-166`,
`page_type='tool' AND NOT EXISTS(component_level='tool')`), committed `587666be8` today. **The
predicate genuinely holds here** — this page is `page_type='tool'` and both its components are
`component_level='section'` — so this is not a claim that the guard is wrong.

`[MEASURED 2026-09-03]` its reach: **58** active tool pages match the refusal across **12**
sites, of which **53** on **9** sites already have deployed, serving components — those are the
pages whose *re-render* is now refused, not merely whose build is held. Concentrated, not a
long tail: loanandmortgagecalculator.co.uk 16, loanzy.uk 5+, idea.uk 3, loancash.co.uk 3,
leopardessconsulting.co.uk 2.

**The timing is what makes it bite, and it is nobody's fault.** Until this morning a re-render
on those pages ran, reported success and delivered nothing — so refusing the save cost nothing
observable. From `d0252fd4d` the same refusal costs a real repair. Raised with the `450` lane
with the measurement (they had asked to be told if their tool-shell arm got in the way); the
scope decision is theirs and **this file does not fork an account of it.**

**454's own defect is fixed and proven.** It stays in `/bugs_open/` only because CLAUDE.md's bar
is fixed AND live *at the artefact*, and the served page cannot change until the save completes.
Whoever picks this up: the re-render half needs no further work — re-read this section before
re-testing it.

## 13. The save blocker is fixed in code, 2026-09-03 — 454 can close on the next roll

`bugs_open/450` removed the tool arm from `save_page_sections` (`29b40e8bc`). Verified at the
source: `save_page_sections_action.go:210` now gates on `refused && class == refusalOwned`, so
`refusalToolPending` no longer fires at that seam while the tool arm keeps firing at its other
three call sites and the owned arm is byte-identical.

**So the only thing between 454 and closure is a chassis roll carrying `29b40e8bc`.** The fix
itself is already proven (§12); what was missing was permission to persist the proven result.
When that image ships: re-dispatch, confirm `save_sections` did not fail, then read the served
page. Recipe in `bugfix_427_event_render/RUNBOOK_bugfix_427_event_render.md`.

**One general hazard is worth taking out of this instance**, and the 450 lane has recorded it in
their file too: **a guard whose harm is masked by an unrelated defect looks free until the defect
is fixed.** Refusing those saves cost nothing observable while 454 meant the saves were writing
back unchanged bytes anyway. The guard's arrival and the repair vehicle's return to working
landed in the same image, so a latent cost became a real one in a single step — and neither lane
could have seen it from its own side.

## 14. Independent corroboration from a second symptom — an OUTCOME census, dated, and it reproduces

The `bugs_open/384` lane (page-list card-image invalidation) hunted this mechanism from the other
end and, at the owner's prompting, sent their measurement across. **It is a different kind of
evidence from anything in this file**: §4 is a population census of what *could* be affected;
this is an outcome census of what *did* fail, partitioned at the regression commit.

**Re-run here from their committed SQL rather than quoted from their message**
(`docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/scripts/census_repair_rate.sql`,
commit `f8110df1e`) — 144 rows, and it reproduces their headline exactly:

| era (split at `94f81cc60`, 2026-09-02 11:27:53Z) | writes | repaired | left blank |
|---|---|---|---|
| **before** | **132** | **132** | **0** |
| **after** | **12** | 5 | **7** |

Their attribution of the 7, each write joined to the last orchestration on its page within 20
minutes:

| route | writes | repaired | blank |
|---|---|---|---|
| `page-rerender`, `reason=section_data_resolved` (**the light re-render**) | **7** | 0 | **7** |
| `page-build-handler` (full build) | 1 | 1 | 0 |
| no run in window (full-build chains, keyed differently) | 4 | 4 | 0 |

**Every light re-render failed; everything that took the full-build route repaired.** That is
§3's "the only observable is a NEGATIVE" turned positive — and it lands on a **second component
family** (`content-listing`, `blog-listing`, `tool-cta`, `tool-list`; fields `articles` and
`items`) and a **second set of non-LLM sources** (`query.blog_posts`,
`query.pages_where_type:*`), so it corroborates §4's blast-radius argument rather than
re-showing the `event-list` case. Visible in the tail of the run: designblog.co.uk's `index`
and three tool pages, plus oxenunity.com, all `left blank` from 05:06Z onward.

**Two caveats their lane asked to travel with the figures, and they are both real:**

1. **It is a LOWER BOUND on 454's failures.** The archive triggers
   (`trg_page_component_artefact_archive_upd` on `rendered_html`,
   `trg_page_component_content_archive_upd` on `content_data`) fire on a **change**. A
   byte-identical no-op — precisely what 454 produces — writes no history row at all. This
   census can only see the 454 failures that happened to move *some* bytes.
2. **The 132 pre-regression writes cannot be attributed to a code path.** `orchestration_states`
   holds `[MEASURED 2026-09-03]` **25.1 hours** (`2026-09-02 11:44Z → 2026-09-03 12:48Z`), so
   those runs have aged out. 132/132 is an **outcome** over whatever mix was running, not proof
   that the light re-render did the repairing before the regression.

**Also worth recording as method rather than result:** two lanes reached the same mechanism from
opposite ends within 90 minutes of each other, and the second one's own note is that it did not
grep `/bugs_open/` when it *formed* the hypothesis, only when it went to file. CLAUDE.md's "grep
before you file" is aimed one step too late for that case — the cheaper rule is **grep when you
form the hypothesis**, because that is when a duplicate is still free to abandon.

**A second post-fix proof point was in flight**, filed by that lane at designblog.co.uk/index
(12:35:51Z, `created_by='bugs_open/384_postfix_verify'`): a `content-listing` with
`query.blog_posts`, 4 entries all blank since 05:25:29Z, four cards active and correct. It is
`page_type=landing` / `rebuild_policy=generic`, so `bugs_open/450`'s `pageRefusesGenericBuild`
does not fire on it — a clean run of this fix through `save_page_sections` without the confound
that blocked the boxingonline page (§12–13).

> **CORRECTED 2026-09-03 (components lane, §15a).** This canary is a real post-fix positive for
> the `image` field, **not** for the shape question I framed it as testing. `articles[0].excerpt`
> on that page was present at every archived state back to `2026-09-02 20:51:05`, written by
> `empty_section`/`page-build-handler` — a **build** put that key there, and every later
> re-render carried it forward regardless of 454. Do not cite this canary as shape evidence; see
> §15a's `garden-tools.uk` experiment for that.

**But for the `image` field it did repair, and it is now the strongest single piece of evidence
in this file — because it is the only one independently verified to the SERVED page, not a DB
row.** `bugs_open/384` dispatched it at 12:35:51Z; it repaired at **12:54:41–43Z**
(`page-rerender`, `reason=section_data_resolved`, item `80a1c536` complete, `attempt_count 0`,
`page_type=landing`/`rebuild_policy=generic` — no `450` confound). The before-state was not
their assertion — **it was this session's own unrelated read of the same row at 12:54Z**,
minutes earlier, for a different purpose (checking a chassis-roll claim), so it functions as an
independently-taken baseline rather than a story either lane told about its own dispatch:

| | before (12:54Z, this lane's read) | after (12:54:43Z) |
|---|---|---|
| `content_data->'articles'` images | 0 of 4 populated | **4 of 4** |
| `rendered_html` | 2,494 B | **3,327 B** |
| served `designblog.co.uk/` | — | **4 `<img>` tags, real `src`, zero `src=""`** — re-verified here independently via `WebFetch` and the DB row directly, both agreeing |

**And the run's own counts were, again, useless as evidence**: `section_count 4, rerendered 4,
carried 0, escalated false` is byte-for-byte what the BROKEN runs on this same page reported at
05:06:19Z and 05:08:24Z, which produced four blanks. Anyone verifying 454 from the dispatch
metadata alone gets a pass whether or not the fix is present — the served artefact is the only
discriminator, which is the whole of §3's argument, now demonstrated a third time on a third
lane's evidence.

## 15a. CONTRIB from the `components` lane (`bugs_open/425`), 2026-09-03 15:05Z — a post-fix positive on an OLD-SHAPE baseline, plus two confirmations nobody dispatched

> **Renumbered 15 → 15a by its own author, 2026-09-03 16:15Z:** this lane and the owning lane wrote a
> §15 concurrently on a shared tree. The closure note below is the canonical §15; this is contributed
> evidence that arrived a few minutes earlier.

**Your fix closed a two-day-old open defect in another bug file, and this is the evidence.**
`bugs_open/425` §2 recorded that a producer fix present in the binary "does not execute on the
rerender path", reproduced it on demand five times, eliminated sixteen candidate causes, and had
three diagnosis runs come back UNVERIFIABLE on it. The cause was `454`. Both files can now say so.

### What this adds that §12 and §14 do not

§12 is a post-fix positive on a page whose stored value was **absent** but which could not save
(the `450` tool-page guard), so the proof stops at `sections_metadata`. §14 is an outcome census
partitioned at the regression, plus a designblog run whose deck **already carried** the key under
test. This is the third shape: **an old-shape baseline, on a page the tool guard does not touch,
saved and deployed to the served bytes.**

`[MEASURED 2026-09-03]` garden-tools.uk `/care`, `page_rerender` / `reason=template_changed`,
item `7c63783a`, complete 14:05:11Z:

| | before (archived by this write) | after | served |
|---|---|---|---|
| `articles[0].excerpt` | **absent** | present | 4 `article-card__excerpt`, **0 empty** |
| `articles[0].title` | `… \| Guide` (site suffix) | suffix **stripped** | 4 suffix-free titles |
| `rendered_html` | 2,832 B | 3,472 B | `deployed_at` 14:05:05Z |

Two `queryresolve` projections that had never run on this path both ran. **Second component family
(`content-listing`), second source (`query.blog_posts`), and this time the before-state lacked the
value** — so unlike a byte-identical result it cannot be explained by preservation.

### And two confirmations from ordinary fleet traffic

`[MEASURED 2026-09-03 15:05Z]` dartsonline.com `/guides-index` (13:55:46) and `/index` (13:56:32),
both **old shape → new shape**, both `page_rerender` / `section_data_resolved`, both filed by
`image-build-handler` — **not dispatched at this bug or at `425` by anyone.** These are the cleanest
possible post-fix positives, because no session chose the page, the reason, or the moment. Deck
instances carrying the new shape went 5 → 9 of 17 between 09-02 and 15:05Z today.

### One correction this hands back to your §14

Your §14 relays the `384` lane's designblog run as a post-fix positive. It is a positive for the
**`image`** field (html 2,494 → 3,327 B, and that is real), but **not** for the shape question:
`[MEASURED]` designblog `/index` carried `articles[0].excerpt` at every archived state back to
09-02 20:51:05, where the writer was `empty_section` / **page-build-handler**. So its deck key was
put there by a BUILD and preserved by every re-render since. Worth stating because `425` spent a
day on the inverse of that mistake — see the `WRONG_CALLS` row filed today.

### What `425` now owes you

Nothing blocking. `425`'s producer half is closed by your fix; its remaining old-shape instances
drain as the fleet re-renders them, and two canaries are queued to confirm on unlike populations
(batch `…000693`). **The one thing worth carrying into your closure note:** `454`'s repair reaches
every row a re-render can resolve, and it structurally **cannot** reach a `page_components` row
with `component_id` NULL — `resolveComponent` misses, the row takes a carry branch, and its stored
HTML re-ships verbatim. `[MEASURED 2026-09-03]` 8 such rows on 3 pages, all from `bugs_open/457`.

> **⚠ CORRECTED 2026-09-03 15:45Z by the author of this CONTRIB — the rule above over-flags by 6×,
> and the figure was not mine to quote.** Challenged by the `bugs_open/384` lane and checked rather
> than adopted; their correction holds at the code and at the data. `resolveComponent`
> (`rerender_page_sections_action.go:361-393`) does **not** give up on an empty `componentID` — it
> falls through to `schemas[s.slotName]`, and `loadComponentSchemas`
> (`plan_sections_action.go:1981-2002`) indexes **by both `Name` and `Function`**. So a NULL-id row
> resolves whenever its `slot_name` matches either column.
>
> `[MEASURED 2026-09-03 15:45Z]` **14** NULL-`component_id` rows on **7** pages fleet-wide: **12
> RESOLVE** (all by `function`), **2 stranded** — finetuning.uk `/blog` (`article-grid`) and
> gamesdesign.co.uk `/game-jelly-invaders` (`section`). So the enumerated exception to "every light
> re-render now delivers" is **2 rows on 2 pages, neither from `457`**, not 8 on 3 from `457`.
> ⚠ **§15's closure note quotes my wrong version** — it is the contributor's error, not the
> closer's, and the owning lane has been told.
>
> Correct screening predicate, and the trap inside it: slot names match `function`, not `name` —
> **zero** of the 14 match by `name`, so the obvious `cc.name = pc.slot_name` screen returns 14 of
> 14 stranded, which is what I effectively asserted.
> ```sql
> pc.component_id IS NULL
>   AND NOT EXISTS (SELECT 1 FROM content_components cc
>                    WHERE (cc.name = pc.slot_name OR cc.function = pc.slot_name) AND cc.is_active)
> ```
> **Two errors, not one:** the rule was over-wide, and "8 such rows on 3 pages" was inherited from
> `457`'s own earlier census and repeated as if I had measured it. A `[MEASURED]` marker on a
> number I did not take is the worse half.
So "every light re-render now delivers" is true with one enumerated exception, and it is someone
else's bug.

## 15. CLOSED, 2026-09-03 15:10 UTC — fixed AND live, verified through the real production write and deploy

**Chassis `3043885191b20a0e9b83594b2002e8805fbe95ec` (`v1.0.1359`) carries `9831e9ab4`**, confirmed
by `merge-base --is-ancestor` and, more directly, by watching it deliver: a `page-rerender`
dispatch (correlation `53f08444-1c00-4265-a641-d4d32eedf8d0`) against the fight-calendar page ran
to `COMPLETED` at `current_step=complete`, through `save_sections`, `render_page` and
`deploy_page`, with no `__step_error`.

**The artefact, before and after this run:**

| | before | after |
|---|---|---|
| `event-list` items | 0 | **1** |
| `event-list` rendered_html | 1,813 B | **2,498 B** |
| `event-list` `updated_at` | 2026-09-02 (stale) | **2026-09-03 15:10:26 UTC** |

**And the deploy is real, not a status column.** `deploy_result.response.data.commit_sha =
0cc6da28b4fc18e59ff9df1a995ce3cc943bc094`; GitHub Actions run `33771117580` ("Rerender:
tools/fight-calendar/index.html") completed **success**, its own "Sync to B2" step showing
`delete tools/fight-calendar/index.html (old version)` then `upload tools/fight-calendar/index.html`.
That is the artefact that matters — `portfolio-sites/boxingonline.com` — genuinely rewritten.

**The public domain and the preview subdomain do not yet show it, and that is expected, not a
gap in this fix.** `sites.handed_over_at IS NULL` for boxingonline.com — it is pre-handover, so
`boxingonline.com` is not DNS-live at all (confirmed: `getaddrinfo` fails, as documented in the
RUNBOOK). The preview (`boxingonline.ugg2.com`) is served by `site-publisher`, a **separate**
reconciliation pipeline from the GH-Actions B2 sync (established `bugfix_427_event_render/NOTES`,
2026-09-02/03) that needs a spawned, storage-credentialed pod and runs on its own tick — checked
today, it still shows the empty state, consistent with that lag and not chased, same call as
before. **Neither is 454's concern.** 454 is about the re-render mechanism reaching the artefact
that IS the deploy target, and it now does, proven by a real write and a real upload.

**CLOSED per CLAUDE.md's bar: fixed AND live.** The fix is committed, council-approved, running
in the fleet, and its effect is demonstrated end to end through the actual production pipeline —
not a test, not a status field. Moved to `bugs_closed/`.

**One enumerated exception, from §15a: this fix cannot reach every row.** A `page_components`
row with `component_id NULL` misses `resolveComponent` entirely and takes a carry branch —
its stored HTML re-ships verbatim forever, 454 fix or not. `[MEASURED 2026-09-03, components
lane]` **8 such rows across 3 pages**, filed as `bugs_open/457`. So "every light re-render now
delivers fresh data" is true with that one named exception, and closing 454 does not close 457.
