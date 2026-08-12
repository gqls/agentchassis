# 266 — an `archived` page is rebuilt and re-stamped `deployed` by at least four independent producers, none of which reads `pages.status`

**Filed 2026-08-12** by the `bugs_open/215` quiet-mode lane, closing that file's §"Two
defects found while doing this" item 1 and the 090 loose end in
`brochure_component_library/HANDOFF_2026-08-12_215_quiet_mode_continue_here.md` §7.

**Status: OPEN. The damage is LIVE and RECURRING — the most recent re-deploy was
today, 2026-08-12 14:25Z.** Nothing has been changed in the tree by this filing.

## The symptom, measured

Two pages on `fundamentallyai.com` were hand-archived on 2026-08-08 by the
fundamentallyai sweep front (`deployed_at IS NULL`, zero components). They are still
`status='archived'` and they are serving:

```
 name                       | status   | build_status | deployed_at
 tool-llm-cost-calculator   | archived | deployed     | 2026-08-11 19:05:36.493684+00
 ai-readiness-checker-guide | archived | deployed     | 2026-08-12 14:25:21.495362+00
```

Artefact check, 2026-08-12 (a fabricated URL is the control, so the check can come out
negative):

```
200  25861b  https://fundamentallyai.com/blog/ai-readiness-checker-guide.html
200  35331b  https://fundamentallyai.com/tools/llm-cost-calculator/index.html
404   2697b  https://fundamentallyai.com/blog/definitely-not-a-real-page-control.html
```

Note the `deployed_at` values: **they are not the 08-11 10:34 / 11:13 stamps the 215
file recorded.** Both pages have been re-deployed *again* since, the latest four hours
before this filing. This is not a historical incident with a residue; it is a loop that
is still running.

## Root cause — it is not one missing predicate, it is four producers and no shared gate

The 090 diagnosis (`38099787-c7f9-46d4-b75e-3a1867fcaf41`, 5 bundle iterations,
2026-08-11 13:03–13:25) reached this, and **I re-verified every claim of it first-hand
against `site_work_items` before recording it here**:

| work item | producer (`created_by`) | `reason` | completed | matching `deployed_at` |
|---|---|---|---|---|
| `2840024e` `needs_page` | `reconcile_site_plan` | `not_built` | 08-11 10:34:24.51 | 10:34:21.99 ✓ |
| `051d46eb` `owned_page_review` | `reconcile_site_plan` | `not_built` | **never — `needs_human_review`** | — |
| `ac21f8d6` `needs_page` | **`image-build-handler`** | `image_landed` | 08-11 11:13:39.01 | 11:13:25.64 ✓ |
| `75981275` `page_rerender` | `completeness-discovery-agent` → `page-rerender` | `cta_links_stale` | 08-11 19:05:44.05 | 19:05:36.49 ✓ |
| `851f7114` `section_edit` | `claude-ideauk-copy-20260812` → `section-editor` | — | 08-12 14:25:31.77 | 14:25:21.50 ✓ |

**The load-bearing observation is the second row.** For `tool-llm-cost-calculator`,
`ReconcileSitePlanAction` did *not* emit a build — it correctly routed the page to
`owned_page_review` / `needs_human_review`, where it still sits, uncompleted. The gate
worked. `image-build-handler` then rebuilt and deployed that same page **sixteen minutes
later** through a completely unrelated path (`needs_imagery` → image lands → `needs_page`
with reason `image_landed`).

So the mechanism the 215 file recorded — "plan still names the page → reconcile emits
`needs_page` → build → deploy", PLAN-017's documented regeneration trap — is **real but
accounts for only one of the two pages**, and for neither of the two subsequent
re-deploys. A fix in `reconcile_site_plan_action.go` would have closed the one door that
was already shut and left the three that were actually used standing open.

> **This corrects the 215 file's own correction.** That file talked itself *down* from
> "distinct defect" to "the documented regeneration trap, not a new one". The
> regeneration trap explains page one; three further producers explain the rest. The
> narrowing was wrong in the safe-looking direction.

### What no reader checks

- `loadRealisedPages` (`platform/orchestration/actions/reconcile_site_plan_action.go:459-464`)
  — `SELECT ... FROM pages WHERE site_id = $1`, no status predicate. [VERIFIED, read]
- `UpdatePageStatusAction` (`platform/orchestration/actions/v3_site_actions.go:648`), deploy
  branch at `:866-874` — `UPDATE pages SET build_status=$2, deployed_at=NOW(), ... WHERE id = $1`.
  No status predicate. It is the **only** writer of `pages.deployed_at` in the estate
  (`grep -rn "deployed_at = NOW()" --include=*.go platform/ internal/` returns this line and
  two `sites` writers). [VERIFIED]
- That function already carries three refusals — `pageHasComponents`, `pageSectionShortfall`,
  and the assembly-skip refusal added by `bugs_open/210` (fixed, live `v1.0.1268`). **None
  reads `pages.status`.** So the idiom and the place for a refusal both already exist.

## Fix candidates, ordered by what they make unrepresentable

1. **Refuse at the commit seam (`deploy_page` / `git_commit`,
   `platform/orchestration/actions/git_deployer_actions.go`).** For an archived page there
   is **no legitimate deploy path at all**, so the correct seam is the one the owned-page
   guard deliberately avoids. Closes all four observed producers and any fifth.
2. **Refuse at `assemble_page`, copying `owned_page_guard`'s placement — DO NOT DO THIS
   BY REFLEX; it closes only two of the four doors.** `owned_page_guard.go:29-36` states
   why it chose `assemble_page`: `git_commit` "is also how owned pages LEGITIMATELY
   deploy", because `page-rerender` (`rerender_single_page`) and `section-editor`
   (`apply_section_edit`) commit pages without passing through `assemble_page`. Those two
   are precisely the producers behind the 19:05 and 14:25 re-deploys above. **`archived`
   is not the same shape as `owned`** — owned means "not the generic pipeline's to
   rebuild", archived means "nothing may deploy this", and that difference moves the seam.
   [Placement rationale INHERITED from that file's doc comment, measured 2026-08-06, not
   re-measured by me — re-measure which loops call `assemble_page` before building.]
3. A predicate in each producer — N doors, and the fifth producer arrives unguarded.
4. "Operators must retract after archiving" — this is the current de facto state and it is
   a defect, not a remedy. See the `098` note below.

## Relation to `bugs_closed/098` — this defeats a closed bug's remedy

`098` ("archiving a page does not retract it from the deployed site") was closed
2026-08-06 with population zero. Its resolution was deliberate: archiving does **not**
auto-retract; a two-step `page-retraction` procedure is the mechanism. That is not what
is failing here.

What this bug shows is that **the retraction primitive is not durable against a
rebuild.** 098's acceptance test was "0 new `page_rerender` rows for any retracted page
since dispatch", measured over pages that happened to have no active producers. These two
pages have four. Retract them today and the next `section_edit`, image landing, discovery
sweep or reconcile pass puts them back. **098 should not be reopened on this file's
evidence** — its own population is still zero — but anyone relying on retraction being
durable should read this first.

## How to verify a fix

1. Pick an archived page with an open producer; do **not** use a page with no work items,
   which is the shape 098's acceptance accidentally measured.
2. Induce each of the four paths and assert the page stays `status='archived'`,
   `deployed_at` unchanged, and the URL 404s — **with a live control page that must still
   deploy in the same run**, or a guard that refuses everything passes.
3. Standing detector for the class, no fix required to run it:

```sql
SELECT s.domain, p.name, p.status, p.build_status, p.deployed_at
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status = 'archived' AND p.deployed_at IS NOT NULL
ORDER BY p.deployed_at DESC;
```

**Population is NOT yet measured fleet-wide.** The 215 file records open work items
sitting on archived pages across **8 domains**, so this is very unlikely to be a
fundamentallyai quirk, but that figure counts work items, not re-deployed archived pages,
and it is not the same measurement. Run the query above before quoting a scope.

## Provenance

- 090 diagnosis run `38099787-c7f9-46d4-b75e-3a1867fcaf41`; artefacts in
  `diagnosis_artifacts` (`kind='bundle'`, iterations 1–5), **expire 2026-09-10, unpinned**.
  Its findings are reproduced above so this file does not depend on them surviving.
- Every table row quoted here was re-queried first-hand 2026-08-12 ~19:00Z.
