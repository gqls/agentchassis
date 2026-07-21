# Handoff — for a DEPLOYED page, `sections=[]` means "rendered elsewhere", not "awaiting composition" — and the planner does not know the difference

> ## ✅ CLOSED 2026-07-21 — FIXED AND LIVE on v1.0.1146
>
> The §4 fix is implemented, verified, and running in production. The reconcile
> logic now gates the empty-sections case on **deployed-ness**, exactly as §3/§4
> prescribe, in both Pass B and Pass B2.
>
> **What shipped** (`platform/orchestration/actions/v3_site_actions.go`,
> `reconcilePlanWithRealised`):
> - **Pass B** — realised page reached by URL under a different name: carry the
>   realised sections, EXCEPT `!deployed && empty` → take the LLM's sections
>   (a catalogued first-plan page is finally composable).
> - **Pass B2** — realised page reached by name: `non-empty` → restore realised;
>   `empty && deployed` → **force the LLM's proposal back to empty** (closes the
>   pre-existing injection exposure §4 names); `empty && !deployed` → keep the
>   LLM's proposal (fall through).
> - New log line `validate: forced deployed sectionless page back to empty` is the
>   discriminating marker for the Pass B2 branch.
>
> **Verification**
> - 4 discriminating unit tests in `v3_site_reconcile_test.go`
>   (`TestReconcile_DeployedEmptyPageStaysEmpty_PassB{,2}`,
>   `TestReconcile_NotDeployedEmptyPageTakesLLMSections_PassB{,2}`): all pass, and
>   neutralising the two *new* gates fails exactly the two new-behaviour tests
>   (`_PassB2` deployed-force-empty, `_PassB` not-deployed-take-LLM) while the two
>   invariant-guard tests keep passing — the §"How to verify" #1/#2 method.
> - **Live**: the discriminating string `forced deployed sectionless page back to
>   empty` is present in the running `agent-chassis:v1.0.1146` pod binary
>   (`strings /app/agent-chassis | grep -c` = 1), positive control present. The
>   fix is a chassis Go change, so being in the running image is what makes it live.
>
> **How it shipped — a process note.** The change was authored on a thread working
> this bug, then swept into the `v1.0.1146` build commit `fe2ba5e52`
> ("sweep … several bugfixes") together with unrelated `bugs_open/037` and
> `bugs_open/041` work — the `git add -A` hazard CLAUDE.md documents. Nothing was
> lost (the committed code is byte-identical to the authored fix and to the §4
> plan), but it means the change reached production **without** the council review
> §5 recommended. The evidence in §2 is strong (fleet-wide measurement + control
> group) and the fix is strictly *more* conservative for deployed pages, so the
> live risk is low; the recommended review, if run, is now retrospective and would
> spawn a follow-up bug only on REVISE/REJECTED.
>
> **Residual / adjacent items unaffected by this closure:** the duplicate-page
> observation in §6 (tool subsystem, not the planner) is still just recorded, not
> filed. `bugs_open/037` (needs_rebuild membership) is a *different* thread's work
> and interacts with this gate — see the `TestReconcile_NeedsRebuildEmptyPageIsStillComposable`
> test note: needs_rebuild joins the preserved-set *membership* but the empty-gate
> deliberately still keys on `realisedPageIsBuilt` (== deployed only), so a
> needs_rebuild page with empty sections stays composable.
>
> Everything below is the original filing, kept verbatim as the diagnosis record.

---

**Filed 2026-07-20**, splitting the surviving residual out of `/bugs_open/001` so that case can close.
**Read the correction in §2 before implementing anything**: 001's own prescription for this residual
is unsafe as written, and applying it would create an injection risk of exactly the class 001 exists
to prevent.

## 1. What 001 left behind

`/bugs_open/001` closed with a "Known residual": Pass B (the URL-match rename snap-back in
`reconcilePlanWithRealised`, `v3_site_actions.go`) replaces the LLM page with
`normaliseRealisedToPlanPage(rp)` **wholesale**, carrying the realised `sections` including an empty
`[]`, and `continue`s before Pass B2's non-empty gate can run. Its prescription:

> The fix would be to keep the realised identity but take the LLM's sections when the realised ones
> are empty — i.e. give Pass B the same gate — but that changes which fields Pass B is allowed to
> carry, so it wants its own review rather than a quiet widening.

It also mis-states the reachable set. It describes the victim as *"a catalogued (uncomposed) page
whose URL the LLM reuses under a different name"*. A merely-catalogued page cannot reach Pass B:
`existingPages = preserved` (`:4601`) filters to adoption-locked **or** deployed pages *before*
`realisedByURL` is built, so Pass B only ever sees preserved pages.

## 2. CORRECTION — the prescribed fix is unsafe, because empty does not mean uncomposed

Measured fleet-wide, 2026-07-20 — every active, deployed page with no sections:

```sql
SELECT s.domain, p.name, p.page_type,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id) AS n_components
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status='active' AND p.build_status='deployed'
  AND jsonb_array_length(coalesce(p.sections,'[]'::jsonb))=0;
-- 18 rows: 14 page_type='tool', 2 'blog-index', 2 'content'
-- 15 of the 18 have n_components >= 1
```

**They render content while carrying no sections.** They are composed by a path other than the
section composer (the tool generator; the blog-index renderer). `sections=[]` on a deployed page is
a positive statement — *this page is not section-composed* — not an absence waiting to be filled.
`idea.uk/tool-audience-check` is the explicit case: URL `/tools.html#audience-check`, an anchor into
another page, deliberately made a sectionless pointer page as the documented idea.uk workaround.

So 001's prescription — "take the LLM's sections when the realised ones are empty" — would let a
re-plan attach a generic section layout (`hero`, `features`, …) to 18 pages that another subsystem
renders. That is content injection onto built pages: the failure class 001 was written to stop.

**Control check, because "tool pages are externally rendered" is too clean to trust.** 6 of 19
deployed tool pages *do* carry sections, so the rule is not by page type:

| domain | name | sections |
|---|---|---|
| ai-agent-orchestration.com | `password-entropy` | `["tool-password-entropy"]` |
| finetuning.uk | `llm-cost-calculator` | `["tool-llm-cost-calculator","tool-cta"]` |
| leopardessconsulting.co.uk | `ai-agent-roi-estimator` | `["tool-ai-agent-roi-estimator","tool-cta"]` |

Their sections are **tool-specific components**, not a generic layout, and their names lack the
`tool-` prefix that all 13 sectionless ones carry. [INFERRED] these are two generations of tool page
— an older one composed through the section composer, a newer one rendered directly — inferred from
the naming and section shape, not from reading the generator. It does not change the conclusion:
either way, what a deployed page's empty `sections` records is "not composed here".

## 3. The genuine defect, and where it actually lives

The original second defect is real: a page preserved with `sections=[]` that genuinely *is* awaiting
composition can never be filled, because the emptiness is carried forward every run. But that page
is **not deployed** — it is preserved because it is adoption-locked, which per `/bugs_open/051`
means the site is on its **first plan**. Every case where the empty-gate was observed helping was a
non-deployed page: dartsonline `guides-index`, `brands-index` and `shop-index` were all `planned`
when they were composed (see 001's live-verification sections).

That splits cleanly:

| realised page | `sections=[]` means | correct behaviour |
|---|---|---|
| `build_status='deployed'` | rendered by another subsystem | **keep empty** — do not let the LLM compose it |
| not deployed (adoption-locked) | catalogued, never composed | **take the LLM's sections** |

## 4. Proposed fix — and note it also closes a live exposure in Pass B2

Gate the empty case on **deployed-ness**, not merely on emptiness, in both passes:

- **Pass B**: when the realised sections are empty *and the realised page is not deployed*, keep the
  realised identity but take the LLM's `sections`. When it is deployed, carry the emptiness (current
  behaviour, now for a stated reason).
- **Pass B2**: today, realised-empty falls through and the LLM's page is kept **as proposed**, so a
  re-plan can already compose a deployed sectionless tool page if it proposes one under the same
  name. Apply the same rule: realised deployed + empty → force `sections` back to empty.

**Pass B2's exposure is pre-existing, not a 001 regression.** Before 001 the preserved set was empty
on every re-plan, the whole function returned `llmPages` untouched, and the LLM's sections were taken
regardless. 001 did not introduce it; it made the guard the place where it can now be closed.

## 5. Why this is not being shipped on one thread's read

The change is small but it is a **behavioural rule about what a re-plan may write to a built page,
fleet-wide**, resting on an interpretation of what an empty `sections` column means. That
interpretation is well-evidenced (§2) but it is an interpretation. Per CLAUDE.md this is the
file-before-you-assert case, and the council gate is the review 001 itself never got (both its
rounds were voided by `/bugs_closed/019`, now closed, so the gate is available again).

Route: council gate on the §4 plan, or a `needs_diagnosis` run on "what does an empty sections column
mean for a deployed page" if a second opinion on §2 is wanted first. **Check `/bugs_open/043` before
relying on the diagnosis loop — diagnosis runs were reported hanging at the route step on 2026-07-20.**

## 6. Adjacent finding, NOT chased here

`ai-agent-orchestration.com` carries the same tool as two deployed pages — `llm-cost-calculator`
(sections, no `tool-` prefix) and `tool-llm-cost-calculator` (sectionless, prefixed) — and likewise
`roi-estimator`/`tool-ai-agent-roi-estimator` and `ai-readiness-quiz`/`tool-ai-readiness-quiz`. That
is a duplicate-page problem belonging to the tool subsystem, not to the planner. Not filed; recorded
so the next thread does not read it as evidence for this bug.

## How to verify a fix

1. Discriminating unit tests in `v3_site_reconcile_test.go` (7 cases already there, all from idea.uk
   fixtures). Two new: deployed+empty must stay empty through **both** Pass B and Pass B2;
   not-deployed+empty must take the LLM's sections through both.
2. Neutralise the new gate and confirm exactly those tests fail — the existing suite was verified
   discriminating this way and the practice should hold.
3. Live: a re-plan on a site with a deployed sectionless tool page (`robot-hands.com/tool-matchmatrix`,
   `vonc.com/tool-arena`) must leave `pages.sections` empty for it, while a `planned` empty page on
   the same run is still composed.
4. **The tree does not currently compile** — another thread has
   `discovery_checks/check_news_feed.go:582` mid-edit and `actions` imports it. Test via
   `git archive HEAD` with your files overlaid; do not edit their file.

## Key references

- `platform/orchestration/actions/v3_site_actions.go` — `reconcilePlanWithRealised` (:4575),
  preserved-set filter (:4601), Pass B (:4698), Pass B2 (:4709), `normaliseRealisedToPlanPage` (:4487),
  `realisedPageIsBuilt` (:4757), `realisedSectionsOf` (:4767).
- `/bugs_open/001` — the residual this splits out; **its prescription is superseded by §2 here**.
- `/bugs_open/051` — why "adoption-locked" means "first plan only", which §3 depends on.
- `/bugs_open/037` — the `needs_rebuild` boundary, a different slice of the same guard.
- `/bugs_closed/019` — why 001 never got a council review.
