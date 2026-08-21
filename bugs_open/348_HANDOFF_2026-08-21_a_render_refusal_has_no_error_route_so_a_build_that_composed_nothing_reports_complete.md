# 348 — a component render REFUSAL has no error route, so a build that composed nothing reports `complete`

**Filed 2026-08-21** by the `mortgagecalculator_couk_adoption` lane. **Status: OPEN, UNOWNED.**
Live on `mortgagecalculator.co.uk` today; **[UNMEASURED] fleet-wide** — see §6.

> **On the 090 loop, stated plainly because the 2026-07-31 owner ruling requires it.**
> This is a cross-cutting structural claim, so it was filed to the diagnosis loop **first**
> (intake `47a4d1d5-5fa3-4940-8d95-f431d5896cb2`, run `0b498cf8-73ac-4d34-9a14-89a84f4e7b7a`).
> **The run FAILED for infrastructure reasons and returned no verdict** — its `verdict` step hit
> `AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 … status 400`, and then its
> own `mark_failed` step failed too (§5, a second defect). So the loop neither confirmed nor
> refuted this; **it never reached the question.**
> I am substituting first-hand verification and naming it: **two independent A/B runs on one
> pinned item**, the **live routing tables read from `agent_definitions` for both agents**, and the
> **guard code read at the deciding arm**. Every claim below is one of those three. Nothing here is
> inferred from the loop.

## 1. The one-paragraph version

`page-build-handler` has two hand-written guard steps whose own descriptions say they exist to
*"park the work item visibly instead of letting the dispatch loop stamp it complete"*. They cover
two no-op causes. **`bugs_closed/260`'s fix (live 2026-08-20) created a third cause — a component
render *refusal* — and there is no guard for it.** The refusal fails a step inside
`page-content-writer`'s `process_sections_loop`, which declares no `error_step`; the failure
returns to `page-build-handler`'s `spawn_content_writer`, which also declares no `error_step`; the
saga reaches its success-labelled complete path and the dispatch loop stamps the item `complete`.
**A build that wrote nothing reports success, and 260's excellent named diagnosis lands in a
terminal item's `error` column where nothing looks.**

## 2. Evidence — two runs on one item, pre-state pinned

Item `0c65f9fa-ddce-4e83-a6a8-4f252b3cf3cb` (`content_rewrite`, site
`62b5978e-4271-4589-8e00-4baebfc0447c`, page `scorecard-simulator`).

**Pre-state, pinned before touching anything** (10:28:27Z): `status='needs_human_review'`,
`attempt_count 0`, error `step validate_content failed: … 20 blockers, 0 errors`, page
`build_status='planned'`, **0 `page_components`**.

| | attempt 1 (10:32Z) | attempt 2 (11:06Z) |
|---|---|---|
| terminal `status` | **`complete`** | **`complete`** |
| `attempt_count` after | 1 | 2 |
| `page_components` | **0** | **0** |
| `pages.build_status` | `planned` | `planned` |
| live URL | 404 | 404 |
| error hash | `24859342` | `62b415f3` (**differs — a fresh error, not the retained one**) |

Both errors:

```
step process_sections_loop_iter_1_render_section failed: failed to execute action
render_component: component "mechanism-flow": content does not match the declared field
type(s) — steps[N].branches: declared array (items: object), got string;
steps[M].branches: … refusing to render (bugs_open/260) (code: CHILD_ORCHESTRATION_FAILED)
```

with **`N,M` = 2,3 on attempt 1 and 1,2 on attempt 2** — which is how the second error is known to
be freshly produced rather than left over.

**The `status='complete'` outcome reproduced on both runs.** The orchestration chain for attempt 1:
the render orchestration **FAILED**, its parent reached `complete_error`, and the outer sagas
reported `complete`/`COMPLETED`.

## 3. Why — read from the live agent definitions, not inferred

```sql
SELECT k, v->>'error_step' FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
WHERE a.type IN ('page-build-handler','page-content-writer')
  AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;
```

| agent | step | `error_step` |
|---|---|---|
| `page-content-writer` | `process_sections_loop` (owns `…_iter_N_render_section`) | **(none)** |
| `page-build-handler` | `spawn_content_writer` | **(none)** |
| `page-build-handler` | `validate_content` | `mark_needs_review` ← **the PRE-260 path** |
| `page-build-handler` | `mark_no_ready_sections` | `complete_error` |
| `page-build-handler` | `mark_writer_skipped` | `complete_error` |

The two `mark_*` steps carry these descriptions **in the live row**:

> *"No ready sections — park the work item visibly instead of letting the dispatch loop stamp it
> complete"*
> *"Writer skipped — park the work item visibly instead of letting the dispatch loop stamp it
> complete"*

`CompleteWorkItemAction` (`platform/orchestration/actions/load_work_item_actions.go`, the `UPDATE …
WHERE status NOT IN ('needs_human_review','failed',…)` arm) **cannot help**: its own comment says it
exists to preserve *"a status a handler deliberately set"*. Nothing flagged this item, so the guard
has nothing to preserve.

**So the defect is a routing gap, not a bookkeeping bug:** the pre-260 failure went through a
*routed* step (`validate_content → mark_needs_review`) and was visible; the post-260 failure goes
through an *unrouted* one.

## 4. Why this is not `bugs_closed/028`, and why that matters

`028` ("a page-build no-op reported `complete`") closed 2026-07-25 — and its fix **is**
`mark_no_ready_sections` above. Same shape, different cause. **The per-cause guard pattern does not
survive a new cause**, and 260 added one three days ago. This is `bugs_open/328`'s own argument
(*"the route needs one rule, not one rule per cause"*) arriving on a different route, which is an
argument for fixing it once at the seam rather than adding a third `mark_*` step.

⚠ **The attempt ladder is bypassed, which is subtler than it looks.** `attempt_count` DID increment
(0→1→2 of 3). But `complete` is terminal, so the item never retries itself and never reaches
`failed`; attempt 2 happened only because a human re-armed it. A counter that advances while the
status short-circuits it reads, in any census, as "the retry machinery is working".

## 5. Second defect found in passing: the failure ladder's own SQL is broken

The 090 run's `mark_failed` step failed with:

```
step mark_failed failed: failed to execute action fail_work_item:
  failed to apply work item failure ladder: ERROR: could not determine data type of parameter $4 (SQLSTATE 42P18)
```

**So when a run fails, the machinery that records the failure also fails.** Untyped `$4` in the
failure-ladder statement — a cast is the likely fix. Not chased here; it needs its own file and is
noted so the next reader does not attribute a missing failure record to the run never happening.

## 6. Fleet incidence — MEASURED 2026-08-21 11:10Z, and the figure is YOUNG, not low

**The new cause has fired exactly ONCE fleet-wide, and that once is this file's own item.**

```sql
SELECT status, count(*), count(DISTINCT site_id)
FROM site_work_items WHERE error LIKE '%refusing to render%' GROUP BY 1;
--  complete | 1 | 1
```

⚠ **Do not read that as "rare".** `bugs_closed/260`'s fix went live 2026-08-20 14:45Z, so the
refusal has existed for about **twenty hours**, and the one firing was a build a human re-armed by
hand. This census cannot yet distinguish *"the shape is uncommon"* from *"the shape has barely had
a chance to occur"*, and it will not be able to for several days. **Re-run it before sizing this.**
It does at least discriminate: it found the known row, so a zero would have meant something.

**The population that stands to shift is not small.** The path 260 moved the failure *off* is
`validate_content`, and it carries real volume:

```sql
SELECT status, count(*) FROM site_work_items
WHERE error LIKE '%validate_content%' AND updated_at > '2026-08-01' GROUP BY 1;
--  needs_human_review | 124
--  complete           |  26
--  cancelled          |  12
```

**124 items were parked VISIBLY since 2026-08-01 by the routed path.** [INFERRED, not measured]
only the mistyped-LLM-field *subset* of those would now arrive via the unrouted path instead —
260's own census put its template-leak class at 26 events across 7 domains, so the subset is a
fraction of 124, not all of it. **Do not quote 124 as the blast radius.** The honest statement is:
the routed path is heavily used, and this defect removes the routing from part of it.

⚠ **Note the 26 `complete` in that same comparator.** The complete-on-error shape appears to occur
on the `validate_content` path too, which this file has NOT investigated — it may be a different
cause wearing the same shape, or evidence the guard is leakier than §3 implies. **[UNVERIFIED.]**

**Still not measured:** whether other handlers that spawn a writer share the routing gap.

## 7. Fix candidates, ordered by what makes the bad state unrepresentable

1. **A step that fails must not be able to reach a success-labelled complete.** Make the absence of
   an `error_step` on a step that can fail a *definition-time* error, or default it to
   `mark_item_failed` rather than to nothing. Kills the class, including causes not yet invented.
2. **Gate completion on the artefact, not the saga.** The completion verifier
   (`complete_work_item_verification.go`) already exists and already fails CLOSED since RFC_017 —
   register a verifier for the page-building item types asserting `page_components` is non-empty.
   Cheaper than 1 and reuses machinery built for exactly this.
3. **Add a third `mark_render_refused` guard.** Matches the existing pattern, fixes this instance,
   and leaves the fourth cause to be discovered the same way. Weakest, and §4 is the argument.

## 8. How to verify a fix

Re-arm item `0c65f9fa-ddce-4e83-a6a8-4f252b3cf3cb` (it reproduces on demand — twice, above) and
require the terminal status to be **not** `complete`. **Positive control in the same run**: a build
that genuinely succeeds must still reach `complete`, or the fix has merely stopped completing
things. Assert at `page_components` and the served URL, never at the item status — the item status
is the thing under test.

## 9. Where the record lives

`docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/NOTES_mortgagecalculator_couk.md`,
`## 2026-08-21`. Related: `bugs_closed/260` (created the cause; its writer half is
`copy_quality_two_stage`'s), `bugs_closed/028` (same shape, earlier cause), `bugs_open/328` (the
same "one rule, not one per cause" argument), `bugs_open/033` (where parked items go to be unread).
