# Handoff — a page flagged `needs_rebuild` still loses its composition to a re-plan

**Filed 2026-07-20**, measured while verifying the `/bugs_open/001` fix live. This is a **boundary
decision left open by that fix**, not a regression it introduced — the behaviour is the same as
before 001; 001 simply protected `deployed` pages and left this case exactly where it was.

**It may well be the wanted behaviour.** It is filed so the choice is made deliberately rather than
inherited from an implementation detail.

## What happens

`/bugs_open/001`'s guard preserves a realised page's composition when it is adoption-locked **or
built**, where "built" is `build_status = 'deployed'`:

```go
// v3_site_actions.go
func realisedPageIsBuilt(rm map[string]interface{}) bool {
	status, _ := rm["build_status"].(string)
	return status == "deployed"
}
```

`needs_rebuild` is therefore outside the preserved set, so a re-plan takes the LLM's proposed
composition for such a page — even though the page may have a full, previously-deployed composition
sitting in `pages.sections`.

## Evidence (dartsonline.com, 2026-07-20, plan `fba367c9` → `5d438145`)

`index` was `needs_rebuild` with seven sections. The re-plan replaced them:

| | sections |
|---|---|
| before (realised) | `hero, category-listing, product-grid, differentiators, call-to-action, testimonials, content-listing` |
| LLM proposed | `hero, product-grid, category-listing, features, call-to-action, testimonials` |
| plan took | the LLM's — **unchanged from the proposal** |

Net: **lost `differentiators` and `content-listing`, gained `features`**, and two slots reordered.
These are distinct components, not naming variants — checked against `content_components`, unlike
the `about` case on the same run (see `/bugs_open/039` Part 1 for that trap):

```sql
SELECT name, function FROM content_components
WHERE function IN ('features','content-listing','differentiators');
--  features                | features
--  content-listing         | content-listing
--  differentiators-section | differentiators
```

In the same run, `contact` was also `needs_rebuild` and came through unchanged — but only because
the LLM happened to re-propose its exact composition, **not** because anything protected it. So the
exposure is real and its effect is a coin-flip on what the LLM writes that run.

## The argument each way

**For leaving it as-is.** `/bugs_open/001`'s own fix step 4 proposed precisely this as the escape
hatch: *"Consider gating a deliberate rebuild behind explicit intent (a per-page `rebuild:true` in
the `needs_site_plan` spec, or **a page whose `build_status` was set to `needs_rebuild`**), so a
genuine redesign is still possible — just never the silent default."* On that reading `needs_rebuild`
IS the explicit intent, and the current behaviour is the design working. It also keeps a way to
redesign a page at all, which a blanket guard would remove.

**Against.** `needs_rebuild` is set by machinery, not only by a human asking for a redesign —
`v3_site_actions.go:644` sets `build_status='needs_rebuild', built_from_plan_version=NULL` as part of
ordinary status handling. So a page can acquire the flag for reasons that have nothing to do with
wanting a new design, and then silently lose its composition at the next re-plan. "Rebuild this
page" and "recompose this page from scratch" are different intents sharing one flag. Fleet-wide this
is not a corner case:

```sql
SELECT rebuild_policy, build_status, count(*) FROM pages GROUP BY 1,2;
--  generic | needs_rebuild | 27
--  owned   | needs_rebuild |  7
```

34 pages currently sit in this state.

## Fix candidates (if the decision is to change it)

1. **Separate the two intents.** Keep `needs_rebuild` meaning "re-render this page as planned"
   (composition preserved, like `deployed`), and introduce an explicit `needs_replan` /
   per-page `rebuild: true` in the `needs_site_plan` spec for "recompose it". This is fix step 4
   done properly and makes the silent case impossible.
2. **Preserve unless the page's composition is empty.** Widen `realisedPageIsBuilt` to
   `deployed OR needs_rebuild`, relying on Pass B2's existing non-empty gate to let a genuinely
   uncomposed page still be composed. One-line change, but it removes the only current route to a
   deliberate redesign — do not take it without (1) or some other way back in.
3. **Surface rather than decide.** When a re-plan would change a `needs_rebuild` page's composition,
   emit the diff as a review item instead of applying it silently. Correct in spirit, but
   `/bugs_open/033` says that queue currently has no working surface, so it would rot — fix that
   first or this is a no-op.

Recommend **(1)**, and until it exists, treat `needs_rebuild` as "this page's composition is
unprotected" when planning any re-plan.

## How to verify a fix

1. Set a deployed page to `needs_rebuild` without asking for a redesign, re-plan, and assert its
   `pages.sections` is unchanged.
2. Assert a page with genuine redesign intent (whatever (1) settles on) IS still re-composed — do
   not fix this by making redesign impossible.
3. Check the artefact: the rebuilt page's `page_components` should match the preserved list, matching
   on `function` per `/bugs_open/039` Part 1.

## Related

- `/bugs_open/001` — the fix this sits at the edge of; its "VERIFIED LIVE" section records this as
  limit 1 and points here.
- `/bugs_open/038` — the other half: even a *protected* page is still rebuilt and its content
  regenerated.
- `/bugs_open/039` — why the `differentiators` / `differentiators-section` distinction above needed
  checking before this could be called a real loss.
- `/bugs_open/033` — why fix candidate 3 would currently rot.
- `/bugs_open/050` — the deployed-empty classification; my fix composes with it (see below).

---

## RESOLVED 2026-07-21 — decision made (candidate 2), fix LIVE on v1.0.1146

**The "leave as-is" reading is REFUTED by the code, so this is a real defect, not the wanted
behaviour.** The "For leaving it as-is" argument above rests on `needs_rebuild` being the explicit
redesign intent (fix step 4's escape hatch). It is not. Every writer of `needs_rebuild` **preserves
`pages.sections`** and means *"re-render this page as planned"*, never *"recompose it from scratch"*
(grep of all setters, 2026-07-21):

- `v3_site_actions.go:644` `UpdatePageStatusAction` — refuses a 0-component (and, via the in-flight
  040 work, a partial) deploy; sets `needs_rebuild`, clears `built_from_plan_version`, **keeps
  sections**.
- `maintenance_actions.go` `flagPagesForRebuild` — an image/maintenance rebuild; keeps sections.
- `store_generated_component_action.go` `markPagesForRebuild` **and**
  `discovery_checks/check_unresolved_sections.go` — flag a page `needs_rebuild` **precisely because
  its `sections` already name a component that has just become available**. The sections are the
  whole point; recomposition would *defeat* the rebuild. These two make the "redesign intent"
  reading impossible.

**Fleet state when fixed** (live DB, 2026-07-21): 26 active `needs_rebuild` pages — 19 carry a real
composition (the protected case), 5 empty (2 dartsonline `section-index` awaiting composition, 3
robot-hands `tool` pages rendered elsewhere per 050; 1 more content page):

```
 page_type    | empty/has_sections | count
 blog-index   | has_sections       |  1
 blog-post    | has_sections       |  5
 content      | has_sections       |  9
 landing      | has_sections       |  2
 section-index| empty              |  2   <- awaiting composition (brands-index, shop-index; 0 comp)
 tool         | empty              |  3   <- rendered elsewhere (robot-hands; 1 comp each), 050's case
 tool         | has_sections       |  4
```

**Decision: candidate 2, done so it composes with the in-flight 050 work.** Introduced a SEPARATE
membership predicate `realisedPageCompositionIsPreserved(rm)` = `deployed OR needs_rebuild`, used at
the two preservation-**membership** sites (the preserved-set filter `:4728` and the truncation
must-keep `:2800`). The empty-sections classification in Pass B/B2 **deliberately stays on
`realisedPageIsBuilt` (= `deployed`)**, because a `needs_rebuild` empty page may be *awaiting
composition* (dartsonline `brands-index`), not *rendered-elsewhere* — Pass B2's existing non-empty
gate routes both kinds correctly.

**Why a separate predicate and not just widening `realisedPageIsBuilt`:** the 050 work overloaded
that predicate for the empty-gate too. A naive widening would **force-empty** an awaiting-composition
`needs_rebuild` page. Proven discriminating in an isolated worktree:

| test (`v3_site_reconcile_test.go`) | without fix | naive widening | this fix |
|---|---|---|---|
| `NeedsRebuildPageCompositionSurvivesReplan` | FAIL | pass | pass |
| `NeedsRebuildPageOmittedByLLMIsUnioned` | FAIL | pass | pass |
| `RealisedPageCompositionIsPreserved` | FAIL | pass | pass |
| `NeedsRebuildEmptyPageIsStillComposable` | pass | **FAIL** | pass |

**Landed & LIVE.** The Go change was swept into the `v1.0.1146` fleet build (owner sweep commit
`fe2ba5e52`, "sweep. v3 site actions … several bugfixes"). Tests committed separately (`9864fab37`,
this session). Verified live on the running pod `agent-chassis-55bbccfdbc-xrkv6`:
`strings /app/agent-chassis | grep -c realisedPageCompositionIsPreserved` = **1** (positive control
`reconcilePlanWithRealised` = 2; negative control = 0). Whole fleet is on `v1.0.1146`.

**The redesign route (candidate 2's caveat, answered).** A built page can no longer be *silently*
redesigned by a re-plan. To deliberately recompose one, empty its `pages.sections`; Pass B2's
non-empty gate then lets the LLM compose it. (For a `deployed` page, 050 makes `deployed`+empty
authoritative, so the deliberate route runs through `needs_rebuild`+empty.) So the "only route to a
redesign" the handoff warned candidate 2 removes is **not** removed — it just becomes an explicit
"empty the composition" act instead of a silent side effect of a status flag.

### Two things left OPEN (neither blocks closing this defect)

1. **Candidate 1 (explicit redesign intent) is a deferred FEATURE, not required to fix this bug.**
   A clean `rebuild:true` / `needs_replan` per-page spec field (fix step 4) would replace the
   "empty the composition" dance with an explicit signal. `/bugs_open/001` deferred fix step 4 as
   "a policy call, not a bug fix"; that call is still open. **Owner decision.**
2. **Live behavioural verification not yet fired.** Unit tests are discriminating and the symbol is
   live, but a re-plan on a site carrying a `needs_rebuild` page (e.g. dartsonline `contact`:
   `needs_rebuild`, 3 sections) to watch `pages.sections` survive has not been run — it mutates a
   live site + ~30 min dispatch + build spend. [UNVERIFIED — live re-plan]

## CLOSED — 2026-07-22 (owner ruling)

Moved to `/bugs_closed/`. The bar is **fixed AND live**, and the headline defect meets it: a re-plan
can no longer silently re-compose or drop a `build_status='needs_rebuild'` page. Fixed via
`realisedPageCompositionIsPreserved` (`v3_site_actions.go`, swept into `v1.0.1146`), tests
`9864fab37`, symbol verified in the running pod. The two open items were put to the owner
(2026-07-22):

1. **Live re-plan verification — owner ruled: close on current evidence.** The discriminating unit
   tests + live pod-grep are the accepted bar (the same machinery `/bugs_closed/001` verified live
   twice). The live re-plan step is written down in the workstream RUNBOOK if ever wanted.
2. **Candidate 1 (explicit redesign intent) — owner ruled: BUILD IT.** This is a follow-on **feature**,
   not part of this defect: a per-page `recompose_pages` signal in the `needs_site_plan` spec so a
   deliberate single-page redesign is a clean flag instead of the "empty the composition, then
   re-plan" dance. Tracked separately in **`/features_open/012`** so this bug can close on its own
   terms. Candidate 1 does not reopen 037 — 037's guard stands; 012 adds an *explicit* opt-out to it.

Workstream docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_037_needs_rebuild_guard/`.
